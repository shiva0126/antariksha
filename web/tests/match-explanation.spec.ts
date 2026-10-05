import {test,expect} from '@playwright/test';

test('chart comparison explains every factor and labels unavailable AI',async({page,baseURL})=>{
 const headers={Origin:baseURL!};
 expect((await page.request.post('/api/auth/register',{headers,data:{email:`matching_${Date.now()}@example.com`,password:'matching-browser-password',birth_date:'1996-05-14',birth_time:'10:15',consent:true}})).status()).toBe(201);
 try {
  await page.goto('/#match');
  await page.getByRole('button',{name:'Match charts',exact:true}).click();
  const report=page.getByRole('region',{name:'Detailed compatibility explanation'});
  await expect(report).toBeVisible();
  await expect(report.locator('details')).toHaveCount(9);
  await expect(report.getByText(/not a percentage chance/)).toBeVisible();
  await report.locator('summary').filter({hasText:'Nadi'}).click();
  await expect(report.locator('details').filter({has:page.locator('summary').filter({hasText:'Nadi'})}).getByText('For this comparison:',{exact:true})).toBeVisible();
  await expect(report.getByText(/cannot diagnose health, genetics or fertility/)).toBeVisible();
  await expect(page.getByRole('heading',{name:/traditional points — not a relationship verdict/})).toBeVisible();
  await page.getByLabel('Birth date',{exact:true}).first().fill('1994-11-03');
  await expect(report).toHaveCount(0);
  await expect(page.getByRole('region',{name:'Matching result',exact:true})).toHaveCount(0);
  await page.getByLabel('Use AI for a more detailed explanation').check();
  await page.getByRole('button',{name:'Match charts',exact:true}).click();
  await expect(report.getByRole('status')).toContainText(/AI is unavailable|AI-assisted|did not pass checks/);
 } finally {await page.request.delete('/api/me',{headers});}
});

test('matrimony explains missing preferences and stops after a block',async({page,browser,baseURL})=>{
 const headers={Origin:baseURL!},other=await browser.newContext({baseURL});
 async function setup(request:typeof page.request,suffix:string){
  expect((await request.post('/api/auth/register',{headers,data:{email:`compare_${Date.now()}_${suffix.replaceAll(' ','_')}@example.com`,password:'comparison-browser-password',birth_date:'1996-05-14',birth_time:'10:15',consent:true}})).status()).toBe(201);
  expect((await request.put('/api/community/settings',{headers,data:{community:true,birth_date:'1996-05-14',avatar:'sun',interests:['hiking']}})).status()).toBe(200);
  expect((await request.put('/api/matrimony/me',{headers,data:{active:true,consent:true,details:{display_name:suffix,city:'Bengaluru',min_age:18,max_age:70}}})).status()).toBe(200);
  return (await request.get('/api/me')).json();
 }
 await setup(page.request,'Viewer');const peer=await setup(other.request,'Comparison profile');
 try {
  await page.goto('/#matrimony');
  await page.getByRole('article',{name:'Comparison profile'}).getByRole('button',{name:'Compare our answers'}).click();
  await page.getByRole('dialog').getByRole('button',{name:'Explain this match',exact:true}).click();
  const report=page.getByRole('region',{name:'Detailed compatibility explanation'});
  await expect(report).toBeVisible();await expect(report.locator('details')).toHaveCount(8);
  await expect(report.locator('summary').filter({hasText:'Family plans'})).toContainText('Not shared by both');
  await report.locator('summary').filter({hasText:'Family plans'}).click();
  await expect(report.locator('details').filter({has:page.locator('summary').filter({hasText:'Family plans'})}).getByText(/not counted as disagreement/)).toBeVisible();
  await report.locator('summary').filter({hasText:/^Location —/}).click();
  await expect(report.getByText(/A city label does not tell you/)).toBeVisible();
  await expect(report.getByText(/same wording/).last()).toBeVisible();
  expect((await other.request.post('/api/community/blocks',{headers,data:{target:(await (await page.request.get('/api/me')).json()).id,block:true}})).status()).toBe(200);
  await page.getByRole('dialog').getByRole('button',{name:'Explain this match',exact:true}).click();
  await expect(report).toHaveCount(0);await expect(page.getByText('profile unavailable',{exact:true})).toBeVisible();
 }finally{await page.request.delete('/api/me',{headers});await other.request.delete('/api/me',{headers});await other.close();}
});
