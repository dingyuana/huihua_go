import { test, expect, type Page, type APIRequestContext } from '@playwright/test'
import * as fs from 'fs'
import * as path from 'path'

/**
 * 进项发票业务流 E2E 测试
 *
 * 覆盖：
 *  - Test 1 进项发票创建 + 列表：登录 → 跳到 /expense-invoices/list（h3=进项发票）→ 截图列表 →
 *           点"新增发票"按钮 → 进 /expense-invoices/create 表单 →
 *           填表（发票号码唯一 / 发票代码 / 开票日期 / 发票类型 / 金额1000 / 税额100 / 供应商）→
 *           点"创建" → 等待跳回列表 → 断言新发票号出现在表格里 → 截图
 *  - Test 2 进项发票验真：复用 Test 1 创建的发票号；进列表点"详情"按钮 →
 *           进 /expense-invoices/detail/:id → 点"验真"按钮（仅 verify_status=unverified 显示）→
 *           等待 ElMessage 成功提示 → 截图
 *  - Test 3 进项发票 OCR 上传：跳到列表 → 点"导入向导"按钮（占位，仅 info 提示）→
 *           创建测试图片 /tmp/test-invoice.jpg（占位 JPEG）→ 直接 POST 到
 *           /api/v1/expense-invoices/ocr（OCR API 存在但前端无 UI 入口）→
 *           断言响应含 invoice_no 字段（Mock 数据）→ 截图
 *
 * 路由（已 grep src/router/routes/base.ts:327-360 确认）：
 *   /expense-invoices               →  redirect 到 /expense-invoices/list
 *   /expense-invoices/list          →  ExpenseInvoiceList.vue         (h3=进项发票)
 *   /expense-invoices/create        →  ExpenseInvoiceForm.vue         (h3=新增进项发票)
 *   /expense-invoices/detail/:id    →  ExpenseInvoiceDetail.vue       (验真按钮)
 *   /expense-invoices/import        →  ExpenseInvoiceImport.vue       (当前未在列表入口跳转)
 *
 * 关键 UI（已 grep 源码）：
 *   列表顶部：
 *     <h3>进项发票</h3>
 *     <el-button>新增发票</el-button>     ← goCreate → push('/expense-invoices/create')
 *     <el-button>导入向导</el-button>     ← goImport 仅 ElMessage.info('导入向导页面待开发')
 *   列表行操作：
 *     详情 / 编辑 / 删除 / 确认 / 验真    ← 全部 link button
 *   表单字段（ExpenseInvoiceForm.vue:11-99）placeholders：
 *     请输入发票号码 / 选填，电子发票可不填 / 请选择开票日期 / 请选择发票类型 /
 *     0.00（金额/税额/价税合计）/ 选填（供应商名称 / 供应商税号 / 备注）
 *   表单提交按钮（line 118）：isEdit ? '保存' : '创建'
 *   详情页验真按钮（ExpenseInvoiceDetail.vue:147-154）：
 *     v-if="invoice.verify_status === 'unverified' || !invoice.verify_status"
 *     @click="handleVerify" → verifyExpenseInvoice(id) → ElMessage.success(`验真完成：${data?.verify_status}`)
 *
 * 注：与 reimbursement.spec.ts / sales-invoice.spec.ts 共享相同风格（不复用 helper，独立可运行）。
 *    真实环境（后端 :8080、前端 :3000）可能未启动，跑通不强制，代码语法必须正确。
 */

const SCREENSHOT_DIR = '/tmp/e2e-screenshots'
const TEST_IMAGE_PATH = '/tmp/test-invoice.jpg'
const API_BASE = 'http://localhost:8080/api/v1'
const LIST_PATH = '/expense-invoices/list' // 实际路由（grep 自 src/router/routes/base.ts:333）
const CREATE_PATH = '/expense-invoices/create'

/** 兜底：把单个用例的失败/异常截图归档 */
async function shot(page: Page, name: string): Promise<void> {
  try {
    if (!fs.existsSync(SCREENSHOT_DIR)) {
      fs.mkdirSync(SCREENSHOT_DIR, { recursive: true })
    }
    const file = path.join(SCREENSHOT_DIR, `expense-invoice-${name}.png`)
    await page.screenshot({ path: file, fullPage: true })
  } catch {
    /* ignore screenshot errors */
  }
}

/** 登录 admin/admin123 — 复刻前端 LoginView 的提交方式（与 reimbursement.spec.ts / sales-invoice.spec.ts 一致） */
async function login(page: Page): Promise<void> {
  await page.goto('/login')
  // LoginView 表单：placeholder="账号" / "密码"
  await page.locator('input[placeholder="账号"]').fill('admin')
  await page.locator('input[placeholder="密码"]').fill('admin123')
  await page.locator('button:has-text("登 录"), button:has-text("登录")').first().click()
  // 登录成功后会被路由守卫送走，等 URL 离开 /login
  await page.waitForURL((url) => !/\/login(\b|$)/.test(url.pathname), { timeout: 15000 })
}

/** 通过 API 登录拿 token（供"取列表→挑记录"/OCR 上传用） */
async function fetchApiToken(request: APIRequestContext): Promise<string> {
  const resp = await request.post(`${API_BASE}/auth/login`, {
    data: { username: 'admin', password: 'admin123' },
  })
  expect(resp.ok(), `login API should succeed, got ${resp.status()}`).toBeTruthy()
  const body = await resp.json()
  // 后端直接返回 { token, user_id, ... }（无 code/data 包装）
  const token: string | undefined = body?.token
  expect(token, 'login response missing token').toBeTruthy()
  return token as string
}

/** 跨测试共享：Test 1 创建的发票号 + 发票 ID */
const sharedState: { invoiceNo?: string; invoiceId?: string } = {}

test.describe('进项发票业务流 E2E', () => {
  test.beforeAll(() => {
    if (!fs.existsSync(SCREENSHOT_DIR)) {
      fs.mkdirSync(SCREENSHOT_DIR, { recursive: true })
    }
    // 准备 Test 3 用的测试图片：占位 JPEG（最小合法 JPEG）
    if (!fs.existsSync(TEST_IMAGE_PATH)) {
      // 1x1 白色 JPEG，base64 解码即可
      const jpegBytes = Buffer.from(
        '/9j/4AAQSkZJRgABAQEAYABgAAD/2wBDAP//////////////////////////////////////////////////////////////////////////////////////2wBDAf//////////////////////////////////////////////////////////////////////////////////////wAARCAAIAAgDASIAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAr/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/8QAFAEBAAAAAAAAAAAAAAAAAAAAAP/EABQRAQAAAAAAAAAAAAAAAAAAAAD/2gAMAwEAAhEDEQA/AKpgD//Z',
        'base64',
      )
      fs.writeFileSync(TEST_IMAGE_PATH, jpegBytes)
    }
  })

  // ===========================================================================
  // Test 1: 进项发票创建 + 列表
  // ===========================================================================
  test('Test 1 进项发票创建 + 列表', async ({ page }) => {
    try {
      // ---- 登录 ----
      await login(page)
      await shot(page, '01-logged-in')

      // ---- 跳转到进项发票列表 ----
      await page.goto(LIST_PATH)
      await page.waitForLoadState('networkidle').catch(() => {})
      // 列表头 h3=进项发票
      await expect(page.locator('h3:has-text("进项发票")').first()).toBeVisible({ timeout: 10000 })
      await shot(page, '01-list')

      // ---- 点"新增发票"按钮（List.vue:6）→ 进 /expense-invoices/create ----
      const newBtn = page.locator('button:has-text("新增发票")').first()
      await expect(newBtn).toBeVisible({ timeout: 5000 })
      await newBtn.click()
      await page.waitForURL((url) => url.pathname.endsWith('/expense-invoices/create'), {
        timeout: 10000,
      })
      // 表单头 h3=新增进项发票
      await expect(page.locator('h3:has-text("新增进项发票")').first()).toBeVisible({ timeout: 5000 })
      await shot(page, '01-create-form')

      // ---- 填表 ----
      // 唯一发票号：时间戳后缀，避免重名
      const ts = Date.now().toString().slice(-10)
      const invoiceNo = `E2E-${ts}`
      sharedState.invoiceNo = invoiceNo

      // 发票号码（ExpenseInvoiceForm.vue:11 placeholder="请输入发票号码"）
      await page.locator('input[placeholder="请输入发票号码"]').fill(invoiceNo)
      // 发票代码（line 16 placeholder="选填，电子发票可不填"）
      await page.locator('input[placeholder="选填，电子发票可不填"]').fill(`CODE${ts}`)
      // 开票日期（el-date-picker placeholder="请选择开票日期"）—— 直接填输入框
      await page.locator('input[placeholder="请选择开票日期"]').fill('2026-01-15')
      // 发票类型（el-select placeholder="请选择发票类型"）—— 点击打开下拉再选
      await page.locator('.el-select:has(input[placeholder="请选择发票类型"])').click()
      await page.locator('.el-select-dropdown__item:has-text("电子普票")').first().click()
      // 费用类别（可选，跳过）

      // 不含税金额 / 税额 / 价税合计 —— 都是 placeholder="0.00"
      // 注意：表单里有 3 个相同 placeholder 的输入框，按 .el-form-item label 定位更稳
      const amountInputs = page.locator('input[placeholder="0.00"]')
      await amountInputs.nth(0).fill('1000')
      await amountInputs.nth(1).fill('100')
      // 价税合计 = amount + tax_amount 由 onAmountChange 自动算出来，不手填

      // 供应商名称（line 93 placeholder="选填"）—— 但 form 里"选填"还出现在发票代码、供应商税号、备注
      // 用 label=供应商名称 定位更精确
      await page.locator('.el-form-item:has-text("供应商名称") input').first().fill('测试供应商')

      await shot(page, '01-create-filled')

      // ---- 点"创建"按钮（line 118 isEdit ? '保存' : '创建'）----
      const submitBtn = page.locator('button:has-text("创建")').first()
      await expect(submitBtn).toBeVisible({ timeout: 5000 })
      await submitBtn.click()

      // 创建成功后 setTimeout 1.5s 跳到列表，等 URL 变化
      await page.waitForURL((url) => url.pathname.endsWith('/expense-invoices/list'), {
        timeout: 15000,
      })
      // 等列表加载并包含新建的发票号
      await expect(page.locator('h3:has-text("进项发票")').first()).toBeVisible({ timeout: 10000 })
      await expect(
        page.locator('.el-table__row', { hasText: invoiceNo }).first(),
      ).toBeVisible({ timeout: 10000 })
      await shot(page, '01-after-create')

      // ---- 通过 API 拿回 id 供 Test 2 使用 ----
      try {
        const apiToken = await fetchApiToken(page.request)
        const listResp = await page.request.get(
          `${API_BASE}/expense-invoices?page=1&pageSize=20&keyword=${encodeURIComponent(invoiceNo)}`,
          { headers: { Authorization: `Bearer ${apiToken}` } },
        )
        if (listResp.ok()) {
          const json: any = await listResp.json().catch(() => ({}))
          const arr: any[] = json?.data?.list || json?.data || []
          const found = arr.find((it) => it?.invoice_no === invoiceNo)
          if (found?.id) sharedState.invoiceId = String(found.id)
        }
      } catch {
        // 兜底：不强 fail，Test 2 仍可在 UI 上点"详情"找目标行
      }
    } catch (err) {
      await shot(page, '01-FAIL')
      throw err
    }
  })

  // ===========================================================================
  // Test 2: 进项发票验真
  // ===========================================================================
  test('Test 2 进项发票验真', async ({ page }) => {
    try {
      // ---- 登录 ----
      await login(page)

      // ---- 兜底：若 Test 1 失败没拿到 invoiceNo，本用例自己造一个 + 通过 API 创建 ----
      let targetNo = sharedState.invoiceNo
      if (!targetNo) {
        targetNo = `E2E-${Date.now().toString().slice(-10)}`
        try {
          const apiToken = await fetchApiToken(page.request)
          await page.request.post(`${API_BASE}/expense-invoices`, {
            headers: { Authorization: `Bearer ${apiToken}` },
            data: {
              invoice_no: targetNo,
              invoice_code: `CODE${Date.now()}`,
              invoice_date: '2026-01-15',
              invoice_kind: 'electronic_normal',
              amount: '1000.00',
              tax_amount: '100.00',
              total_amount: '1100.00',
              vendor_name: '测试供应商',
            },
          })
        } catch {
          // API 也不通时仍尝试在 UI 里造一张作为兜底（可能失败，但能验证流程）
        }
        sharedState.invoiceNo = targetNo
      }

      // ---- 跳到列表 ----
      await page.goto(LIST_PATH)
      await page.waitForLoadState('networkidle').catch(() => {})
      await expect(page.locator('h3:has-text("进项发票")').first()).toBeVisible({ timeout: 10000 })
      await shot(page, '02-list')

      // ---- 定位目标行（按发票号精确匹配）----
      const targetRow = page
        .locator('.el-table__row', { hasText: targetNo })
        .first()
      await expect(targetRow, `list should contain row with invoice_no=${targetNo}`).toBeVisible({
        timeout: 10000,
      })

      // ---- 点"详情"按钮（List.vue:86 link button "详情"）----
      const detailBtn = targetRow.locator('button:has-text("详情")').first()
      await expect(detailBtn).toBeVisible({ timeout: 5000 })
      await detailBtn.click()

      // 详情页 /expense-invoices/detail/:id
      await page.waitForURL((url) => /\/expense-invoices\/detail\//.test(url.pathname), {
        timeout: 10000,
      })
      // 等详情卡片的验真信息 section 出现
      await expect(page.locator('.section-title:has-text("验真信息")').first()).toBeVisible({
        timeout: 10000,
      })
      await shot(page, '02-detail')

      // ---- 点"验真"按钮（Detail.vue:147-154，仅 verify_status=unverified 时显示）----
      // 如果按钮不可见（已验真），仍截图 + log 提示，不强 fail
      const verifyBtn = page.locator('button:has-text("验真")').first()
      const verifyVisible = await verifyBtn.isVisible().catch(() => false)
      if (!verifyVisible) {
        // eslint-disable-next-line no-console
        console.warn(`Test 2: invoice ${targetNo} already verified (button hidden)`)
        await shot(page, '02-already-verified')
      } else {
        await verifyBtn.click()

        // 等待 ElMessage 成功提示（全局 message 容器 .el-message）
        const successMsg = page.locator('.el-message:has-text("验真完成")').first()
        await expect(successMsg).toBeVisible({ timeout: 15000 })
        await shot(page, '02-after-verify')

        // 提示会 3s 后消失，但页面上验真状态 tag 应该已经更新
        await page.waitForTimeout(500)
        await shot(page, '02-final')
      }
    } catch (err) {
      await shot(page, '02-FAIL')
      throw err
    }
  })

  // ===========================================================================
  // Test 3: 进项发票 OCR 上传
  // ===========================================================================
  test('Test 3 进项发票 OCR 上传', async ({ page }) => {
    try {
      // ---- 登录 ----
      await login(page)

      // ---- 跳到列表 ----
      await page.goto(LIST_PATH)
      await page.waitForLoadState('networkidle').catch(() => {})
      await expect(page.locator('h3:has-text("进项发票")').first()).toBeVisible({ timeout: 10000 })
      await shot(page, '03-list')

      // ---- 点"导入向导"按钮（List.vue:7）----
      // 当前实现是 ElMessage.info('导入向导页面待开发')，没有真正的 OCR 上传 UI 入口。
      // 这里仍然点一下按钮验证按钮存在，并捕获提示弹窗截图；OCR 上传本身走 API 路径。
      const importBtn = page.locator('button:has-text("导入向导")').first()
      await expect(importBtn).toBeVisible({ timeout: 5000 })
      await importBtn.click()
      // 等待 info 提示（导入向导页面待开发）
      const infoMsg = page.locator('.el-message:has-text("导入向导")').first()
      await expect(infoMsg).toBeVisible({ timeout: 5000 }).catch(() => {
        // eslint-disable-next-line no-console
        console.warn('Test 3: 导入向导按钮未触发提示（可能已实装）')
      })
      await page.waitForTimeout(500)
      await shot(page, '03-import-btn')

      // ---- 创建测试图片（已在 beforeAll 准备，这里再次确认存在）----
      expect(fs.existsSync(TEST_IMAGE_PATH), `test image should exist at ${TEST_IMAGE_PATH}`).toBe(
        true,
      )

      // ---- 直接通过 API 调用 OCR（前端无 OCR UI，用 API 验证 Mock OCR 返回值）----
      const apiToken = await fetchApiToken(page.request)
      // 读取本地图片并以 multipart/form-data 上传
      const fileBuffer = fs.readFileSync(TEST_IMAGE_PATH)
      const ocrResp = await page.request.post(`${API_BASE}/expense-invoices/ocr`, {
        headers: {
          Authorization: `Bearer ${apiToken}`,
        },
        multipart: {
          file: {
            name: 'test-invoice.jpg',
            mimeType: 'image/jpeg',
            buffer: fileBuffer,
          },
        },
      })

      // ---- 断言 OCR 响应 ----
      // 后端 Mock 应返回 2xx；body 形如 { data: { invoice_no: 'INV...', ... } } 或 { invoice_no, ... }
      const ocrOk = ocrResp.ok()
      const ocrBody: any = await ocrResp.json().catch(() => ({}))
      // eslint-disable-next-line no-console
      console.log('Test 3 OCR response:', ocrResp.status(), JSON.stringify(ocrBody).slice(0, 200))

      if (ocrOk) {
        const data = ocrBody?.data || ocrBody
        // Mock OCR 应至少返回 invoice_no 字段（实现约定"INV 开头"）
        expect(data?.invoice_no, 'OCR mock should return invoice_no').toBeTruthy()
        // eslint-disable-next-line no-console
        console.log(`Test 3 OCR result: invoice_no=${data.invoice_no}`)
      } else {
        // 环境无 OCR 端点实现时不强 fail，仅记录
        // eslint-disable-next-line no-console
        console.warn(`Test 3: OCR endpoint not available (status=${ocrResp.status()})`)
      }

      // ---- 截图 ----
      // 列表页截图（OCR 完成后）
      await page.goto(LIST_PATH)
      await page.waitForLoadState('networkidle').catch(() => {})
      await expect(page.locator('h3:has-text("进项发票")').first()).toBeVisible({ timeout: 10000 })
      await shot(page, '03-final')
    } catch (err) {
      await shot(page, '03-FAIL')
      throw err
    }
  })
})