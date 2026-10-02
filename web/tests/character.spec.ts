import {test,expect} from '@playwright/test';
test('character imports only reviewed interests and keeps source notes private',async({page,baseURL})=>{
 const headers={Origin:baseURL!};
 expect((await page.request.post('/api/auth/register',{headers,data:{email:`character_browser_${Date.now()}@example.com`,password:'character-browser-password',birth_date:'1996-05-14',birth_time:'10:15',consent:true}})).status()).toBe(201);
 try{
  await page.goto('/#community');await page.getByRole('button',{name:'My character',exact:true}).click();
  await page.getByLabel('Character name',{exact:true}).fill('Curious explorer');
  await page.getByLabel('My bio or post text').fill('I enjoy reading. My friend enjoys hiking.');
  await page.getByLabel('I wrote this text and want to use it for my private profile.').check();
  await page.getByRole('button',{name:'Preview mentioned interests'}).click();
  expect((await (await page.request.get('/api/me/character')).json()).revision).toBe(0);
  const imports=page.locator('section').filter({has:page.getByRole('heading',{name:'Bring your own social story'})});
  await imports.getByRole('checkbox',{name:'reading',exact:true}).check();
  await imports.getByRole('button',{name:'Add reviewed note to my draft'}).click();
  await page.getByLabel('Save these details and my chosen source notes privately in my account.').check();
  await page.getByRole('button',{name:'Save private character',exact:true}).click();
  await expect(page.getByText('Private character saved.',{exact:true})).toBeVisible();
  const stored=await (await page.request.get('/api/me/character')).json();
  expect(stored.profile.interests).toEqual(['reading']);expect(stored.profile.sources).toHaveLength(1);
  await page.reload();await page.getByRole('button',{name:'My character',exact:true}).click();
  await expect(page.getByLabel('Character name',{exact:true})).toHaveValue('Curious explorer');
  await expect(page.getByText('instagram · User-provided',{exact:true})).toBeVisible();
  await page.getByRole('button',{name:'Remove source note'}).click();
  await page.getByLabel('Save these details and my chosen source notes privately in my account.').check();
  await page.getByRole('button',{name:'Save private character',exact:true}).click();
  await expect(page.getByText('Private character saved.',{exact:true})).toBeVisible();
  expect((await (await page.request.get('/api/me/character')).json()).profile.sources).toEqual([]);
 }finally{await page.request.delete('/api/me',{headers});}
});
