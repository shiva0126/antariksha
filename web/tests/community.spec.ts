import {test,expect} from '@playwright/test';

test('private inbox supports read, dismiss and invitation revocation',async({page,browser,baseURL})=>{
 const origin=baseURL!;const headers={Origin:origin};
 const other=await browser.newContext({baseURL});const actor=other.request;
 const credentials=(suffix:string)=>({email:`notice_${Date.now()}_${suffix}@example.com`,password:'notification-test-password',birth_date:'1996-05-14',birth_time:'10:15',consent:true});
 expect((await page.request.post('/api/auth/register',{headers,data:credentials('reader')})).status()).toBe(201);
 expect((await actor.post(origin+'/api/auth/register',{headers,data:credentials('sender')})).status()).toBe(201);
 try{
  const me=await (await page.request.get('/api/me')).json();
  await actor.post(origin+'/api/families',{headers,data:{name:'Notification test family'}});
  const groups=await (await actor.get(origin+'/api/families')).json();
  const url=origin+`/api/families/${groups[0].id}/members`;
  expect((await actor.post(url,{headers,data:{action:'invite',handle:me.handle,role:'viewer'}})).status()).toBe(200);
  await page.goto('/#community');
  await page.getByRole('button',{name:'Notifications',exact:true}).click();
  await expect(page.getByText(/invited you to a private family group/)).toBeVisible();
  await page.getByRole('button',{name:'Mark read',exact:true}).click();
  await expect(page.getByRole('button',{name:'Mark read',exact:true})).toHaveCount(0);
  await page.getByRole('button',{name:'Dismiss',exact:true}).click();
  await expect(page.getByText('No notifications yet.')).toBeVisible();
  await actor.post(url,{headers,data:{action:'remove',handle:me.handle}});
  await actor.post(url,{headers,data:{action:'invite',handle:me.handle}});
  await page.getByRole('button',{name:'Refresh notifications'}).click();
  await expect(page.getByRole('button',{name:'Mark read',exact:true})).toBeVisible();
  await actor.post(url,{headers,data:{action:'remove',handle:me.handle}});
  await page.getByRole('button',{name:'Refresh notifications'}).click();
  await expect(page.getByText('No notifications yet.')).toBeVisible();
 }finally{
  await page.request.delete('/api/me',{headers});
  await actor.delete(origin+'/api/me',{headers});await other.close();
 }
});

test('community opt-in, private post, comment and family tree',async({page,baseURL})=>{
 const origin=baseURL!;
 const response=await page.request.post('/api/auth/register',{headers:{Origin:origin},data:{email:`community_${Date.now()}@example.com`,password:'community-browser-password',birth_date:'1996-05-14',birth_time:'10:15',consent:true}});
 expect(response.status()).toBe(201);
 try{
  await page.goto('/#community');
  await page.getByRole('button',{name:'Notifications',exact:true}).click();
  await expect(page.getByText('No notifications yet.')).toBeVisible();
  await page.getByRole('button',{name:'Profile & interests',exact:true}).click();
  await page.getByLabel('Community introduction',{exact:true}).fill('I enjoy the night sky and hiking.');
  await page.getByLabel('Interests, separated by commas').fill('Hiking, astronomy');
  await page.getByRole('checkbox').check();
  await page.getByRole('button',{name:'Save profile & privacy'}).click();
  await expect(page.getByRole('status')).toContainText('Profile and privacy choices saved.');
  await page.getByRole('button',{name:'Feed',exact:true}).click();
  await page.getByLabel('Caption',{exact:true}).fill('My private stargazing journal');
  await page.getByRole('button',{name:'Save privately',exact:true}).click();
  await page.getByRole('button',{name:'My posts',exact:true}).click();
  await expect(page.locator('.social-post')).toContainText('My private stargazing journal');
  await page.getByRole('button',{name:'Comments',exact:true}).click();
  await page.getByLabel('Your comment').fill('A clear sky tonight');
  await page.getByRole('button',{name:'Comment',exact:true}).click();
  await expect(page.locator('.post-comments')).toContainText('A clear sky tonight');
  await page.getByRole('button',{name:'Family',exact:true}).click();
  await page.getByLabel('New family group name').fill('Our private tree');
  await page.getByRole('button',{name:'Create private group'}).click();
  await page.getByRole('button',{name:'Open group'}).click();
  await expect(page.getByRole('img',{name:'Family relationship map'})).toBeVisible();
  await page.getByLabel('Name or private placeholder').fill('Relative A');
  await page.getByRole('checkbox').check();
  await page.getByRole('button',{name:'Add person',exact:true}).click();
  await expect(page.getByRole('img',{name:'Family relationship map'})).toContainText('Relative A');
 }finally{await page.request.delete('/api/me',{headers:{Origin:origin}});}
});
