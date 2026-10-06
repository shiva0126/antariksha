import {test,expect} from '@playwright/test';

test('member changes password and old sessions cannot access the account',async({page,browser,baseURL})=>{
 test.skip(!process.env.ACCESS_TEST_DATABASE_URL,'requires an isolated application');
 if(!new URL(process.env.ACCESS_TEST_DATABASE_URL!).pathname.startsWith('/astrisk_delivery_test_')||!baseURL?.startsWith('http://127.0.0.1:3109'))throw new Error('Refusing account mutation outside delivery test server');
 const email=`security_browser_${Date.now()}@example.com`,password='original-browser-password',next='new-browser-secure-password';
 const headers={Origin:baseURL!},other=await browser.newContext({baseURL});
 try{
  expect((await page.request.post('/api/auth/register',{headers,data:{email,password,birth_date:'1996-01-01',birth_time:'10:15',consent:true}})).status()).toBe(201);
  expect((await other.request.post('/api/auth/login',{headers,data:{email,password}})).status()).toBe(200);
  await page.goto('/#me/security');
  await expect(page.getByRole('heading',{name:'Active sessions',exact:true})).toBeVisible();
  await expect(page.getByRole('button',{name:'Revoke session',exact:true})).toHaveCount(2);
  await page.getByLabel('Current password for account changes',{exact:true}).fill(password);
  await page.getByLabel('New password',{exact:true}).fill(next);
  await page.getByLabel('Confirm new password',{exact:true}).fill(next);
  await page.getByRole('button',{name:'Change password and sign out',exact:true}).click();
  await expect(page.getByRole('heading',{name:'Welcome to Astrisk'})).toBeVisible();
  expect((await other.request.get('/api/me')).status()).toBe(401);
  expect((await other.request.post('/api/auth/login',{headers,data:{email,password}})).status()).toBe(401);
  await page.getByLabel('Email',{exact:true}).fill(email);
  await page.getByLabel('Password',{exact:true}).fill(next);
  await page.getByRole('button',{name:'Sign in',exact:true}).click();
  await expect(page.getByRole('heading',{name:'Active sessions',exact:true})).toBeVisible();
  await expect(page.getByRole('button',{name:'Revoke session',exact:true})).toHaveCount(1);
  await page.setViewportSize({width:390,height:844});
  expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
 }finally{await page.request.delete('/api/me',{headers});await other.close();}
});
