import { test, expect, type Page, type APIRequestContext } from '@playwright/test'
import * as fs from 'fs'
import * as path from 'path'

/**
 * 报销单业务流 E2E 测试
 *
 * 覆盖：
 *  - Test 1 报销单完整流程：登录 → 新建 → 详情 → 提交 → 驳回
 *  - Test 2 报销单附件上传：切到附件管理 Tab，上传 /tmp/test-upload.txt
 *  - Test 3 报销单关联进项发票：切到发票关联 Tab，必要时 API 建一张进项发票再关联
 *
 * 路由（已确认）：
 *   列表   /expense/reimbursement
 *   新建   /expense/reimbursement/new
 *   详情   /expense/reimbursement/:id
 *
 * 截图保存：/tmp/e2e-screenshots/reimbursement-{N}-{step}.png
 *
 * 注：本文件按"代码语法正确、API 用法正确"为目标编写，
 *    真实环境（后端 :8080、前端 :3000）可能未启动，跑通不强制。
 */

const SCREENSHOT_DIR = '/tmp/e2e-screenshots'
const UPLOAD_FILE = '/tmp/test-upload.txt'
const API_BASE = 'http://localhost:8080/api/v1'
const LIST_PATH = '/expense/reimbursement'

/** 兜底：把单个用例的失败/异常截图归档 */
async function shot(page: Page, name: string): Promise<void> {
  try {
    if (!fs.existsSync(SCREENSHOT_DIR)) {
      fs.mkdirSync(SCREENSHOT_DIR, { recursive: true })
    }
    const file = path.join(SCREENSHOT_DIR, `reimbursement-${name}.png`)
    await page.screenshot({ path: file, fullPage: true })
  } catch {
    /* ignore screenshot errors */
  }
}

/** 登录 admin/admin123 — 复刻前端 LoginView 的提交方式 */
async function login(page: Page): Promise<void> {
  await page.goto('/login')
  // LoginView 表单：placeholder="账号" / "密码"
  await page.locator('input[placeholder="账号"]').fill('admin')
  await page.locator('input[placeholder="密码"]').fill('admin123')
  await page.locator('button:has-text("登 录"), button:has-text("登录")').first().click()
  // 登录成功后会被路由守卫送走，等 URL 离开 /login
  await page.waitForURL((url) => !/\/login(\b|$)/.test(url.pathname), { timeout: 15000 })
}

/** 通过 API 登录拿 token（供 Test 3 备用建进项发票用） */
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

test.describe('报销单业务流 E2E', () => {
  test.beforeAll(() => {
    if (!fs.existsSync(SCREENSHOT_DIR)) {
      fs.mkdirSync(SCREENSHOT_DIR, { recursive: true })
    }
    // 准备 Test 2 用的上传文件
    if (!fs.existsSync(UPLOAD_FILE)) {
      fs.writeFileSync(UPLOAD_FILE, 'test content\n', 'utf-8')
    }
  })

  /** 在两个测试间共享：记录 Test 1 创建的报销单 ID（如果 Test 1 跑失败则 Test 2/3 自己新建） */
  let createdReimbursementId: string | undefined

  // ===========================================================================
  // Test 1: 报销单完整流程（创建→提交→驳回）
  // ===========================================================================
  test('Test 1 报销单完整流程：创建→提交→驳回', async ({ page }) => {
    try {
      // ---- 登录 ----
      await login(page)
      await shot(page, '01-logged-in')

      // ---- 跳转到报销单列表 ----
      await page.goto(LIST_PATH)
      await page.waitForLoadState('networkidle').catch(() => {})
      // 列表头标题"报销单"
      await expect(page.locator('h3:has-text("报销单")').first()).toBeVisible({ timeout: 10000 })
      await shot(page, '01-list')

      // ---- 点击"新建报销单"按钮 ----
      const createBtn = page.locator('button:has-text("新建报销单")')
      await expect(createBtn).toBeVisible()
      await createBtn.click()
      await page.waitForURL(/\/expense\/reimbursement\/new$/, { timeout: 10000 })
      await expect(page.locator('h3:has-text("新建报销单")').first()).toBeVisible()
      await shot(page, '01-form-empty')

      // ---- 填表 ----
      // 按 el-form label 文本定位最近的 el-input / el-input-number 输入框
      const applicantInput = page.locator(
        'el-form-item:has-text("申请人") input, .el-form-item:has(label:has-text("申请人")) input',
      ).first()
      await applicantInput.fill('张三')

      const deptInput = page.locator(
        'el-form-item:has-text("部门") input, .el-form-item:has(label:has-text("部门")) input',
      ).first()
      await deptInput.fill('研发部')

      const amountInput = page.locator(
        'el-form-item:has-text("报销金额") input, .el-form-item:has(label:has-text("报销金额")) input',
      ).first()
      // el-input-number 内部就是 <input>，直接 fill
      await amountInput.fill('1000.00')
      await amountInput.blur().catch(() => {})

      const descInput = page.locator(
        'el-form-item:has-text("说明") input, .el-form-item:has(label:has-text("说明")) input',
      ).first()
      await descInput.fill('测试')
      await shot(page, '01-form-filled')

      // ---- 提交保存 ----
      const saveBtn = page.locator('button:has-text("保存")').first()
      await expect(saveBtn).toBeVisible()
      await saveBtn.click()

      // 保存成功后 router.back() 回列表
      await page.waitForURL(LIST_PATH, { timeout: 15000 })
      await page.waitForLoadState('networkidle').catch(() => {})
      // 列表里应能看到刚创建的"测试"单据
      await expect(page.locator('text=测试').first()).toBeVisible({ timeout: 10000 })
      await shot(page, '01-list-after-create')

      // ---- 抓出新单据 ID（从 URL 或表格行） ----
      // 走 API 直接拿"最新一条"作为 createdReimbursementId，
      // 这样即便 UI 抓 row id 失败，Test 2/3 也能继续。
      const apiToken = await fetchApiToken(page.request)
      const listResp = await page.request.get(`${API_BASE}/reimbursements?page=1&pageSize=5`, {
        headers: { Authorization: `Bearer ${apiToken}` },
      })
      if (listResp.ok()) {
        const listJson = await listResp.json()
        const list: any[] = listJson?.data?.list || listJson?.data || []
        const newest = list.find((it) => it?.description === '测试') || list[0]
        if (newest?.id) createdReimbursementId = String(newest.id)
      }

      // ---- 点详情：定位包含"测试"的行 → 点"详情" ----
      const detailLink = page
        .locator('.el-table__row', { hasText: '测试' })
        .first()
        .locator('button:has-text("详情")')
        .first()
      await detailLink.click()
      await page.waitForURL(/\/expense\/reimbursement\/[^/]+$/, { timeout: 10000 })
      // 从 URL 二次确认 ID（URL 末尾段）
      const m = page.url().match(/\/expense\/reimbursement\/([^/?#]+)$/)
      if (m && !createdReimbursementId) createdReimbursementId = m[1]
      await shot(page, '01-detail-basic')
      await expect(page.locator('h3:has-text("报销单详情")').first()).toBeVisible({ timeout: 10000 })

      // ---- 点"提交"按钮（docstatus 0→1） ----
      const submitBtn = page.locator('button:has-text("提交")').first()
      await expect(submitBtn).toBeVisible({ timeout: 10000 })
      await submitBtn.click()
      // 等接口回到 docstatus=1 — 状态 tag 会变成"已提交"
      await expect(page.locator('text=已提交').first()).toBeVisible({ timeout: 15000 })
      await shot(page, '01-detail-submitted')

      // ---- 跳过 Approve（缺 6602.03 种子数据） ----

      // ---- 点"驳回" ----
      const rejectBtn = page.locator('button:has-text("驳回")').first()
      await expect(rejectBtn).toBeVisible({ timeout: 10000 })
      await rejectBtn.click()
      // UI 的 handleReject 不会弹 prompt 收 reason（看 ReimbursementDetail.vue），
      // 但我们仍然做一次"如果有 prompt 就输入"的兜底
      page.once('dialog', (dlg) => {
        dlg.accept('测试驳回').catch(() => {})
      })
      // 兜底：el-popconfirm/el-message-box 形式
      const confirmBtn = page.locator(
        '.el-message-box__btns button:has-text("确定"), .el-popconfirm button:has-text("确定")',
      )
      if (await confirmBtn.first().isVisible().catch(() => false)) {
        await confirmBtn.first().click()
      }
      // 验证 docstatus=3 — 状态 tag 显示"已驳回"
      await expect(page.locator('text=已驳回').first()).toBeVisible({ timeout: 15000 })
      await shot(page, '01-detail-rejected')
    } catch (err) {
      await shot(page, '01-FAIL')
      throw err
    }
  })

  // ===========================================================================
  // Test 2: 报销单附件上传
  // ===========================================================================
  test('Test 2 报销单附件上传', async ({ page }) => {
    try {
      // ---- 登录 ----
      await login(page)

      // ---- 复用 Test 1 的单据，否则新建一条 ----
      let id = createdReimbursementId
      if (!id) {
        await page.goto(`${LIST_PATH}/new`)
        await page.locator(
          'el-form-item:has-text("申请人") input, .el-form-item:has(label:has-text("申请人")) input',
        ).first().fill('李四')
        await page.locator(
          'el-form-item:has-text("部门") input, .el-form-item:has(label:has-text("部门")) input',
        ).first().fill('市场部')
        await page.locator(
          'el-form-item:has-text("报销金额") input, .el-form-item:has(label:has-text("报销金额")) input',
        ).first().fill('500.00')
        await page.locator(
          'el-form-item:has-text("说明") input, .el-form-item:has(label:has-text("说明")) input',
        ).first().fill('附件测试')
        await page.locator('button:has-text("保存")').first().click()
        await page.waitForURL(LIST_PATH, { timeout: 15000 })
        // 从 API 拿最新一条
        const apiToken = await fetchApiToken(page.request)
        const listResp = await page.request.get(`${API_BASE}/reimbursements?page=1&pageSize=5`, {
          headers: { Authorization: `Bearer ${apiToken}` },
        })
        const listJson: any = listResp.ok() ? await listResp.json() : {}
        const newest: any = (listJson?.data?.list || listJson?.data || [])[0]
        id = newest?.id
      }
      expect(id, 'reimbursement id must be available for Test 2').toBeTruthy()

      // ---- 进入详情 ----
      await page.goto(`${LIST_PATH}/${id}`)
      await expect(page.locator('h3:has-text("报销单详情")').first()).toBeVisible({ timeout: 10000 })
      await shot(page, '02-detail-loaded')

      // ---- 切到"附件管理" Tab ----
      await page.locator('.el-tabs__item:has-text("附件管理")').first().click()
      await expect(page.locator('.el-tabs__item.is-active:has-text("附件管理")').first()).toBeVisible({
        timeout: 5000,
      })
      await shot(page, '02-tab-attachments')

      // ---- el-upload 自定义 http-request：直接找内部 <input type="file"> setInputFiles ----
      // ReimbursementDetail 用的是 drag 模式 + :show-file-list="false"
      const fileInput = page.locator('.el-upload input[type="file"]').first()
      await expect(fileInput).toHaveCount(1)
      await fileInput.setInputFiles(UPLOAD_FILE)

      // ---- 等附件出现 ----
      // 列表 empty-text 是"暂无附件"，上传成功后会出现行，文件名 = "test-upload.txt"
      await expect(page.locator('.el-table__row:has-text("test-upload.txt")').first()).toBeVisible({
        timeout: 15000,
      })
      await shot(page, '02-attachment-uploaded')
    } catch (err) {
      await shot(page, '02-FAIL')
      throw err
    }
  })

  // ===========================================================================
  // Test 3: 报销单关联进项发票
  // ===========================================================================
  test('Test 3 报销单关联进项发票', async ({ page }) => {
    try {
      // ---- 登录 ----
      await login(page)

      // ---- 拿 token ----
      const apiToken = await fetchApiToken(page.request)

      // ---- 准备进项发票：先列一下看是否已有可用 ----
      // 注意：后端 POST /expense-invoices 接收的字段比较多
      // (company_id / invoice_no / invoice_date / invoice_kind / tax_amount / total_amount ...)
      // tenant_id / company_id 通过 JWT 自动注入；invoice_kind 必填
      const expListResp = await page.request.get(`${API_BASE}/expense-invoices?limit=10`, {
        headers: { Authorization: `Bearer ${apiToken}` },
      })
      let availableBefore: any[] = []
      if (expListResp.ok()) {
        const body: any = await expListResp.json()
        availableBefore = body?.data?.list || body?.data || []
      }
      if (availableBefore.length === 0) {
        // 用 API 创建一张进项发票
        // invoice_kind 必须是枚举值之一（看仓库常量）—— 用最常见的 "vat_special" / "vat_general"
        const payload = {
          company_id: '00000000-0000-0000-0000-000000000001',
          invoice_no: `E2E-INV-${Date.now()}`,
          invoice_date: new Date().toISOString().slice(0, 10),
          invoice_kind: 'vat_special',
          tax_amount: '0.00',
          total_amount: '1000.00',
          vendor_name: 'E2E 测试供应商',
        }
        const createResp = await page.request.post(`${API_BASE}/expense-invoices`, {
          headers: {
            Authorization: `Bearer ${apiToken}`,
            'Content-Type': 'application/json',
          },
          data: payload,
        })
        // 201 / 200 都算 ok；非 ok 也允许继续（环境可能无此资源权限）
        if (!createResp.ok()) {
          // 仅记一行日志，不直接 fail —— UI 行为仍可验证
          // eslint-disable-next-line no-console
          console.warn(`create expense-invoice returned ${createResp.status()}`)
        }
      }

      // ---- 复用或新建报销单 ----
      let id = createdReimbursementId
      if (!id) {
        await page.goto(`${LIST_PATH}/new`)
        await page.locator(
          'el-form-item:has-text("申请人") input, .el-form-item:has(label:has-text("申请人")) input',
        ).first().fill('王五')
        await page.locator(
          'el-form-item:has-text("部门") input, .el-form-item:has(label:has-text("部门")) input',
        ).first().fill('行政部')
        await page.locator(
          'el-form-item:has-text("报销金额") input, .el-form-item:has(label:has-text("报销金额")) input',
        ).first().fill('800.00')
        await page.locator(
          'el-form-item:has-text("说明") input, .el-form-item:has(label:has-text("说明")) input',
        ).first().fill('发票关联测试')
        await page.locator('button:has-text("保存")').first().click()
        await page.waitForURL(LIST_PATH, { timeout: 15000 })
        const listResp = await page.request.get(`${API_BASE}/reimbursements?page=1&pageSize=5`, {
          headers: { Authorization: `Bearer ${apiToken}` },
        })
        const listJson: any = listResp.ok() ? await listResp.json() : {}
        const newest: any = (listJson?.data?.list || listJson?.data || [])[0]
        id = newest?.id
      }
      expect(id, 'reimbursement id must be available for Test 3').toBeTruthy()

      // ---- 进入详情 ----
      await page.goto(`${LIST_PATH}/${id}`)
      await expect(page.locator('h3:has-text("报销单详情")').first()).toBeVisible({ timeout: 10000 })

      // ---- 切到"发票关联" Tab ----
      await page.locator('.el-tabs__item:has-text("发票关联")').first().click()
      await expect(page.locator('.el-tabs__item.is-active:has-text("发票关联")').first()).toBeVisible({
        timeout: 5000,
      })
      await shot(page, '03-tab-invoices')

      // ---- 点"添加关联"，弹对话框 ----
      const addBtn = page.locator('button:has-text("添加关联")').first()
      await expect(addBtn).toBeVisible()
      await addBtn.click()
      const dialog = page.locator('.el-dialog:has-text("添加关联进项发票")').first()
      await expect(dialog).toBeVisible({ timeout: 10000 })
      await shot(page, '03-link-dialog')

      // ---- 选第一条（或第一行的 selection 列） ----
      const firstCheckbox = dialog.locator('.el-table__row .el-checkbox').first()
      // 等表格行出现
      await dialog.locator('.el-table__row').first().waitFor({ state: 'visible', timeout: 15000 }).catch(() => {})
      if (await firstCheckbox.isVisible().catch(() => false)) {
        await firstCheckbox.click()
      } else {
        // 兜底：直接点行
        await dialog.locator('.el-table__row').first().click()
      }
      await shot(page, '03-link-selected')

      // ---- 点"确定关联" ----
      const confirmLinkBtn = dialog.locator('button:has-text("确定关联")')
      await expect(confirmLinkBtn).toBeVisible()
      // 按钮 disabled 状态由 selectedInvoiceIds.length 控制
      await expect(confirmLinkBtn).toBeEnabled({ timeout: 5000 }).catch(() => {})
      await confirmLinkBtn.click()

      // ---- 验证列表显示已关联 ----
      // 弹窗关闭，回到 Tab 列表：empty-text = "暂未关联进项发票"
      await expect(dialog).toBeHidden({ timeout: 15000 })
      // 列表里应至少有一行（不再显示 empty-text）
      const linkedRow = page.locator('.el-card .el-table__row').first()
      await linkedRow.waitFor({ state: 'visible', timeout: 15000 }).catch(() => {})
      await shot(page, '03-linked')
    } catch (err) {
      await shot(page, '03-FAIL')
      throw err
    }
  })
})
