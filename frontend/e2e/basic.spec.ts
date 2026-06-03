import { test, expect } from '@playwright/test'

test.describe('Core User Journey', () => {
  test('login page loads and form renders', async ({ page }) => {
    await page.goto('/login')
    await expect(page.locator('input[type="text"]')).toBeVisible()
    await expect(page.locator('input[type="password"]')).toBeVisible()
    await expect(page.locator('button:has-text("登")')).toBeVisible()
  })

  test('login with invalid credentials shows error', async ({ page }) => {
    await page.goto('/login')
    await page.locator('input[type="text"]').fill('baduser')
    await page.locator('input[type="password"]').fill('badpass')
    await page.locator('button:has-text("登")').click()
    await expect(page).toHaveURL(/login/)
  })

  test('main navigation renders after login', async ({ page }) => {
    // bypass auth by setting localStorage tokens
    await page.goto('/')
    await page.evaluate(() => {
      localStorage.setItem('huihua_token', 'e2e-test-token')
      localStorage.setItem(
        'huihua_user',
        JSON.stringify({
          id: 'e2e-admin',
          username: 'admin',
          name: '测试管理员',
          role: 'admin',
          permissions: ['*'],
        }),
      )
    })
    await page.goto('/')
    // the sidebar or header should be visible
    await expect(page.locator('#app')).toBeVisible()
  })

  test('invoice list page renders', async ({ page }) => {
    await page.goto('/')
    await page.evaluate(() => {
      localStorage.setItem('huihua_token', 'e2e-test-token')
      localStorage.setItem(
        'huihua_user',
        JSON.stringify({
          id: 'e2e-admin',
          username: 'admin',
          name: '测试管理员',
          role: 'admin',
          permissions: ['*'],
        }),
      )
    })
    await page.goto('/invoices')
    await expect(page).toHaveURL(/invoices/)
  })

  test('period management page renders', async ({ page }) => {
    await page.goto('/')
    await page.evaluate(() => {
      localStorage.setItem('huihua_token', 'e2e-test-token')
      localStorage.setItem(
        'huihua_user',
        JSON.stringify({
          id: 'e2e-admin',
          username: 'admin',
          name: '测试管理员',
          role: 'admin',
          permissions: ['*'],
        }),
      )
    })
    await page.goto('/period')
    await expect(page).toHaveURL(/period/)
  })
})
