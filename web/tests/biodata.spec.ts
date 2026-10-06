import {test,expect} from '@playwright/test';

test('family prepares photos and biodata, member reviews before publication',async({page,browser,baseURL})=>{
 const headers={Origin:baseURL!}, other=await browser.newContext({baseURL});
 const register=async(request:typeof page.request,suffix:string)=>{const result=await request.post('/api/auth/register',{headers,data:{email:`biodata_browser_${Date.now()}_${suffix}@example.com`,password:'biodata-browser-password',birth_date:'1996-05-14',birth_time:'10:15',consent:true}});expect(result.status()).toBe(201);return (await request.get('/api/me')).json();};
 await register(page.request,'parent');const member=await register(other.request,'member');
 try {
  await page.goto('/#matrimony/family');
  await page.getByText('Prepare biodata for a family member',{exact:true}).click();
  const form=page.locator('form').filter({has:page.getByRole('button',{name:'Create private biodata draft'})});
  await form.getByLabel('Display name',{exact:true}).fill('Family prepared profile');
  await form.getByRole('combobox',{name:'Profile type',exact:true}).selectOption('bride');
  await form.getByLabel('Education',{exact:true}).fill('Engineering');
  await form.getByRole('checkbox').check();
  await form.getByRole('button',{name:'Create private biodata draft'}).click();
  const draft=page.getByRole('article').filter({has:page.getByRole('heading',{name:'Family prepared profile'})});
  await expect(draft).toBeVisible();
  const png=await page.evaluate(()=>{const canvas=document.createElement('canvas');canvas.width=3;canvas.height=3;canvas.getContext('2d')!.fillRect(0,0,3,3);return canvas.toDataURL('image/png').split(',')[1];});
  await draft.getByLabel('Choose profile photo').setInputFiles({name:'portrait.png',mimeType:'image/png',buffer:Buffer.from(png,'base64')});
  await draft.getByLabel('Photo description').fill('Portrait supplied with permission');
  await draft.getByLabel('I have permission from the adult pictured to store this photo.').check();
  await draft.getByRole('button',{name:'Upload photo',exact:true}).click();
  await expect(draft.getByRole('img',{name:'Portrait supplied with permission'})).toBeVisible();
  await draft.getByLabel("Member's Astrisk handle").fill(member.handle);
  await draft.getByRole('button',{name:'Send for their review'}).click();
  await expect(draft.getByText(/Awaiting review/)).toBeVisible();
  const memberPage=await other.newPage();await memberPage.goto('/#matrimony/family');
  await expect(memberPage.getByRole('heading',{name:'Family prepared profile'})).toBeVisible();
  await memberPage.getByRole('button',{name:'Use this biodata privately'}).click();
  await memberPage.getByRole('dialog').getByRole('button',{name:'Confirm'}).click();
  await expect(memberPage.getByText('No drafts or review requests yet.')).toBeVisible();
  await memberPage.getByRole('tab',{name:'My matrimony profile'}).click();
  const own=memberPage.locator('form.mat-editor');
  await expect(own.getByRole('textbox',{name:'Display name',exact:true})).toHaveValue('Family prepared profile');
  await expect(own.getByLabel('Make my profile discoverable to eligible members')).not.toBeChecked();
  await expect(own.getByLabel('Include when I publish')).not.toBeChecked();
  await expect(own.getByRole('img',{name:'Portrait supplied with permission'})).toBeVisible();
 } finally {await page.request.delete('/api/me',{headers});await other.request.delete('/api/me',{headers});await other.close();}
});
