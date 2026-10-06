import {test,expect} from '@playwright/test';

test('private numerology, tropical Western chart and tarot explain their methods',async({page,baseURL})=>{
 const headers={Origin:baseURL!};
 expect((await page.request.post('/api/auth/register',{headers,data:{email:`readings_browser_${Date.now()}@example.com`,password:'readings-browser-password',birth_date:'1996-05-14',birth_time:'10:15',consent:true}})).status()).toBe(201);
 try{
  await page.goto('/#kundali/systems');await expect(page.getByRole('heading',{name:'Other systems'})).toBeVisible();
  await page.getByText('Reading library & AI availability',{exact:true}).click();
  const library=page.locator('details[aria-label="Reading library status"]');
  // Both depend on the server's configuration, so either state is valid.
  await expect(library).toContainText(/AI provider (not )?configured/);
  await expect(library).toContainText(/Semantic search is ready|Semantic enrichment is not ready/);
  await expect(library.getByRole('listitem').filter({hasText:'Lal Kitab (five original volumes)'})).toContainText('Not acquired');
  await page.getByText('Reading library & AI availability',{exact:true}).click();
  await page.getByLabel('Name for numerology (optional)').fill('John');
  await page.getByRole('button',{name:'Read my numbers'}).click();
  await page.getByText('Our numerology convention',{exact:true}).click();
  await expect(page.getByText(/Pythagorean letters A–Z/)).toBeVisible();
  await expect(page.getByText('Expression',{exact:true})).toBeVisible();
  await page.getByRole('button',{name:'Western astrology'}).click();
  await page.getByRole('button',{name:'Read my Western chart'}).click();
  await expect(page.getByRole('heading',{name:/Your Sun, Moon and rising sign/})).toBeVisible();
  await expect(page.getByText(/tropical zodiac/)).toBeVisible();
  await page.getByRole('button',{name:'Tarot',exact:true}).click();
  await page.getByRole('button',{name:'Draw cards'}).click();
  await expect(page.getByText('About this draw',{exact:true})).toBeVisible();
  await expect(page.locator('.tarot-card')).toHaveCount(3);
 }finally{await page.request.delete('/api/me',{headers});}
});
