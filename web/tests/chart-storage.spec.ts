import { test, expect } from '@playwright/test';

test('private charts cross devices with explicit saves and conflict protection', async ({page, browser, baseURL}) => {
 const headers={Origin:baseURL!}, email=`charts_${Date.now()}@example.com`, password='private-charts-test-password';
 const registered=await page.request.post('/api/auth/register',{headers,data:{email,password,birth_date:'1996-05-14',birth_time:'10:15',consent:true}});
 expect(registered.status()).toBe(201);
 const me=await (await page.request.get('/api/me')).json();
 const second=await browser.newContext({baseURL});
 try {
  await page.goto('/#me/charts');
  await expect(page.getByRole('button',{name:'Save charts to my account',exact:true})).toBeVisible();
  const chart={id:'browser-chart',name:'Private browser chart',place:'Bengaluru',date:'1996-05-14',time:'10:15',lat:12.97,lon:77.59,tz:'Asia/Kolkata'};
  const key=`antariksha.member.${me.id}.profiles`;
  await page.evaluate(({key,chart})=>localStorage.setItem(key,JSON.stringify([chart])),{key,chart});
  expect((await (await page.request.get('/api/me/charts')).json()).profiles).toEqual([]);
  await page.getByRole('button',{name:'Save charts to my account',exact:true}).click();
  await expect(page.getByText('Charts saved to your private account.',{exact:true})).toBeVisible();
  expect((await second.request.post('/api/auth/login',{headers,data:{email,password}})).status()).toBe(200);
  const p2=await second.newPage();
  await p2.goto('/#me/charts');
  await expect(p2.getByRole('button',{name:'Save charts to my account',exact:true})).toBeVisible();
  expect(await p2.evaluate(key=>JSON.parse(localStorage.getItem(key)||'[]'),key)).toEqual([chart]);
  await p2.evaluate(({key,chart})=>localStorage.setItem(key,JSON.stringify([{...chart,name:'Second device chart'}])),{key,chart});
  await p2.getByRole('button',{name:'Save charts to my account',exact:true}).click();
  await expect(p2.getByText('Charts saved to your private account.',{exact:true})).toBeVisible();
  await page.getByRole('button',{name:'Save charts to my account',exact:true}).click();
  await expect(page.getByText(/charts changed on another device/)).toBeVisible();
  await page.getByRole('button',{name:'Load account charts',exact:true}).click();
  await page.getByRole('dialog').getByRole('button',{name:'Load charts'}).click();
  await expect(page.getByText('Account charts loaded.',{exact:true})).toBeVisible();
  expect(await page.evaluate(key=>JSON.parse(localStorage.getItem(key)||'[]')[0].name,key)).toBe('Second device chart');
  await page.getByRole('button',{name:'Restore previous device charts',exact:true}).click();
  await page.getByRole('dialog').getByRole('button',{name:'Restore'}).click();
  await expect(page.getByText('Previous device charts restored. Account copy unchanged.',{exact:true})).toBeVisible();
  expect(await page.evaluate(key=>JSON.parse(localStorage.getItem(key)||'[]')[0].name,key)).toBe('Private browser chart');
  expect((await (await page.request.get('/api/me/charts')).json()).profiles[0].name).toBe('Second device chart');
 } finally {
  await page.request.delete('/api/me',{headers});
  await second.close();
 }
});

test('email delivery unavailable is honest and reset links require submission',async({page})=>{
 await page.goto('/');
 await page.getByRole('button',{name:'Forgot password?',exact:true}).click();
 await expect(page.getByText('Email reset is not available yet. Use your saved recovery key.')).toBeVisible();
 let submitted=0;
 await page.route('**/api/auth/email/verify',async route=>{submitted++;await route.fulfill({status:400,contentType:'application/json',body:JSON.stringify({error:'link is invalid, expired or already used'})});});
 await page.goto('/#verify/'+'a'.repeat(64));
 await expect(page.getByRole('button',{name:'Verify my email',exact:true})).toBeVisible();
 expect(submitted).toBe(0);
 await page.getByRole('button',{name:'Verify my email',exact:true}).click();
 await expect(page.getByText('link is invalid, expired or already used',{exact:true})).toBeVisible();
 expect(submitted).toBe(1);
});
