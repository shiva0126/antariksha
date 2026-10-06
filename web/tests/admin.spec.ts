import {test,expect} from '@playwright/test';
import {execFileSync} from 'node:child_process';
import {resolve} from 'node:path';

test('superadmin manages accounts while members cannot access administration',async({page,browser,baseURL})=>{
 const dsn=process.env.ADMIN_TEST_DATABASE_URL;
 test.skip(!dsn,'isolated migrated database required');
 if(!new URL(dsn!).pathname.startsWith('/astrisk_admin_test_'))throw new Error('Refusing admin fixture setup outside isolated test database');
 const headers={Origin:baseURL!},password='admin-browser-test-password',other=await browser.newContext({baseURL});
 const email=`admin_browser_${Date.now()}@example.com`,targetEmail=`target_browser_${Date.now()}@example.com`;
 async function register(request:typeof page.request,mail:string){expect((await request.post('/api/auth/register',{headers,data:{email:mail,password,birth_date:'1996-01-01',birth_time:'10:15',consent:true}})).status()).toBe(201);return (await request.get('/api/me')).json();}
 const owner=await register(page.request,email),target=await register(other.request,targetEmail);
 if(![owner.id,target.id].every(id=>/^[a-f0-9]{64}$/.test(id)))throw new Error('unexpected fixture id');
 const sql=(statement:string)=>execFileSync(resolve('../.runtime/pgdist/usr/lib/postgresql/16/bin/psql'),[dsn!,'-X','-v','ON_ERROR_STOP=1','-c',statement],{stdio:'pipe'});
 sql(`UPDATE member_accounts SET role='superadmin' WHERE id='${owner.id}'`);
 try{
  expect((await other.request.get('/api/admin/users')).status()).toBe(403);
  await page.goto('/#admin');await expect(page.getByRole('heading',{name:'Superadmin',exact:true})).toBeVisible();
  await page.getByRole('textbox',{name:'Search users',exact:true}).fill(targetEmail);
  await page.getByRole('button',{name:'Search users',exact:true}).click();
  await page.getByRole('button',{name:'Manage',exact:true}).click();
  const form=page.getByRole('form',{name:'Manage selected user'});
  await form.getByRole('combobox',{name:'Account action',exact:true}).selectOption('suspend');
  await form.getByLabel('Reason',{exact:true}).fill('Browser suspension check');
  await form.getByLabel('Your current password',{exact:true}).fill(password);
  await form.getByRole('button',{name:'Apply account action'}).click();
  await expect(page.getByText('member · Suspended',{exact:true})).toBeVisible();
  expect((await other.request.get('/api/me')).status()).toBe(401);
  expect((await other.request.post('/api/auth/login',{headers,data:{email:targetEmail,password}})).status()).toBe(401);
  await page.getByRole('button',{name:'Manage',exact:true}).click();
  await form.getByRole('combobox',{name:'Account action',exact:true}).selectOption('reactivate');
  await form.getByLabel('Reason',{exact:true}).fill('Browser reactivation check');
  await form.getByLabel('Your current password',{exact:true}).fill(password);
  await form.getByRole('button',{name:'Apply account action'}).click();
  await expect(page.getByText('member · Active',{exact:true})).toBeVisible();
  expect((await other.request.post('/api/auth/login',{headers,data:{email:targetEmail,password}})).status()).toBe(200);
  await page.getByRole('tab',{name:'Action log',exact:true}).click();
  await expect(page.getByText('Reason: Browser suspension check',{exact:true})).toBeVisible();
  await page.setViewportSize({width:390,height:844});
  expect(await page.evaluate(()=>document.documentElement.scrollWidth<=window.innerWidth)).toBe(true);
 }finally{
  await other.close();
  sql(`DELETE FROM member_audit WHERE actor IN('${owner.id}','${target.id}') OR subject IN('${owner.id}','${target.id}')`);
  sql(`DELETE FROM member_accounts WHERE id IN('${owner.id}','${target.id}')`);
 }
});
