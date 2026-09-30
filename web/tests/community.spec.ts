import {test,expect} from '@playwright/test';

test('community opt-in, private post, comment and family tree',async({page})=>{
 const origin='http://127.0.0.1:3000';
 const response=await page.request.post('/api/auth/register',{headers:{Origin:origin},data:{email:`community_${Date.now()}@example.com`,password:'community-browser-password',birth_date:'1996-05-14',birth_time:'10:15',consent:true}});
 expect(response.status()).toBe(201);
 try{
  await page.goto('/#community');
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
