import {test,expect,type Page} from '@playwright/test';

// Requests go through the page: with COOKIE_SECURE the session cookie is only
// sent by the browser, not by Playwright's API client, over http://127.0.0.1.
async function api(page:Page,method:string,path:string,body?:unknown){
 if(!page.url().startsWith('http'))await page.goto('/');
 return page.evaluate(async([method,path,body])=>{const r=await fetch(path as string,{method:method as string,headers:{'Content-Type':'application/json'},body:body===undefined?undefined:JSON.stringify(body)});return{status:r.status,json:await r.json().catch(()=>null)};},[method,path,body] as const);
}
async function removeAccount(page:Page){expect((await api(page,'DELETE','/api/me')).status).toBeLessThan(300);}

test('Ask Astrisk answers questions about a match, including numerology',async({page,baseURL})=>{
 const headers={Origin:baseURL!};
 expect((await page.request.post('/api/auth/register',{headers,data:{email:`askmatch_${Date.now()}@example.com`,password:'ask-match-browser-password',birth_date:'1996-05-14',birth_time:'10:15',consent:true}})).status()).toBe(201);
 try {
  await page.goto('/#match');
  await page.getByRole('button',{name:'Match charts',exact:true}).click();
  const chat=page.getByRole('region',{name:'Ask Astrisk about this match'});
  await expect(chat).toBeVisible();
  await chat.getByRole('button',{name:'What does numerology say about us?'}).click();
  await expect(chat.locator('.msg-assistant').last()).toContainText('root number (mulank)');
  await expect(chat.locator('.msg-assistant').last()).toContainText('How this lines up with the kundalis');
  await chat.getByLabel('Your question about this match').fill('Explain the Nadi koota');
  await chat.getByRole('button',{name:'Ask',exact:true}).click();
  await expect(chat.locator('.msg-assistant').last()).toContainText('Nadi:');
  await expect(chat.locator('.msg-user')).toHaveCount(2);
  // Follow-up questions keep the topic of the previous one.
  await chat.getByLabel('Your question about this match').fill('tell me more');
  await chat.getByRole('button',{name:'Ask',exact:true}).click();
  await expect(chat.locator('.msg-assistant').last()).toContainText('Nadi:');
  // Changing the details clears the result and its conversation.
  await page.getByLabel('Birth date',{exact:true}).first().fill('1994-11-03');
  await expect(chat).toHaveCount(0);
  // The kundali chat correlates numerology with the chart.
  const r=await api(page,'POST','/api/chat',{birth:{date:'1996-05-14',time:'10:15',lat:12.97,lon:77.59,tz:'Asia/Kolkata'},question:'How does my numerology connect with my chart?'});
  expect(r.status).toBe(200);
  const a=r.json.answer.content as string;
  expect(a).toContain('ruled by Mercury');
  expect(a).toContain('How this connects with your kundali');
 } finally {await removeAccount(page);}
});

test('numerology shows the ruling graha of each root number',async({page,baseURL})=>{
 const headers={Origin:baseURL!};
 expect((await page.request.post('/api/auth/register',{headers,data:{email:`ank_${Date.now()}@example.com`,password:'ank-browser-password-1',birth_date:'1996-05-14',birth_time:'10:15',consent:true}})).status()).toBe(201);
 try {
  await page.goto('/#kundali/systems');
  await page.getByRole('button',{name:'Read my numbers'}).click();
  const card=page.locator('.ank-card');
  await expect(card).toContainText('ruled by Mercury');
  await expect(card).toContainText('ruled by Saturn');
 } finally {await removeAccount(page);}
});
