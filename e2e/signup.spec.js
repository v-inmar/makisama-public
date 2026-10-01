// @ts-check
import { test, expect } from '@playwright/test';


test('signup test', async ({ page }) => {
  await page.goto('http://localhost:5000/auth/signup');
  await page.locator('input[name="firstname"]').click();
  await page.locator('input[name="firstname"]').fill('john');
  await page.locator('input[name="lastname"]').click();
  await page.locator('input[name="lastname"]').fill('doe');
  await page.locator('input[name="email"]').click();
  await page.locator('input[name="email"]').fill('john.doe@email.com');
  await page.locator('input[name="password"]').click();
  await page.locator('input[name="password"]').fill('password');
  await page.locator('input[name="repeat"]').click();
  await page.locator('input[name="repeat"]').fill('password');
  await page.getByRole('button', { name: 'Sign Up' }).click();
  await expect(page.getByText('Hello John! Welcome to Makisama')).toBeVisible();
});