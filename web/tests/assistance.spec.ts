import {test,expect} from '@playwright/test';

test('owner grants family assistance, helper suggests, owner revokes',async({page,browser,baseURL})=>{
 const base=baseURL!,headers={Origin:base};
 const helperContext=await browser.newContext({baseURL}),candidateContext=await browser.newContext({baseURL});
 const helperPage=await helperContext.newPage();
 const requests=[page.request,helperContext.request,candidateContext.request];
 const members:any[]=[];
 try{
  for(let i=0;i<requests.length;i++){
   const r=requests[i];expect((await r.post(base+'/api/auth/register',{headers,data:{email:`helper_${Date.now()}_${i}@example.com`,password:'family-assistance-password',birth_date:'1996-05-14',birth_time:'10:15',consent:true}})).status()).toBe(201);
   members.push(await(await r.get(base+'/api/me')).json());
   expect((await r.put(base+'/api/community/settings',{headers,data:{community:true,birth_date:'1996-05-14',avatar:'sun',accent:'#d6b467',interests:['hiking'],public_bio:'I like hiking'}})).status()).toBe(200);
   expect((await r.put(base+'/api/matrimony/me',{headers,data:{active:true,consent:true,details:{introduction:'Interested in shared adventures',min_age:18,max_age:70}}})).status()).toBe(200);
  }
  await page.request.post('/api/families',{headers,data:{name:'Assistance family'}});
  const families=await(await page.request.get('/api/families')).json();const group=families[0].id;
  await page.request.post(`/api/families/${group}/members`,{headers,data:{action:'invite',handle:members[1].handle}});
  await helperContext.request.post(base+`/api/families/${group}/members`,{headers,data:{action:'accept'}});
  await page.goto('/#community');await page.getByRole('button',{name:'Family assistance',exact:true}).click();
  await page.getByLabel("Helper's account handle").fill(members[1].handle);
  await page.getByLabel('Shared family group').selectOption(group);
  await page.getByRole('button',{name:'Grant shortlist permission'}).click();
  await expect(page.getByRole('button',{name:'Revoke helper'})).toBeVisible();
  await helperPage.goto(base+'/#community');await helperPage.getByRole('button',{name:'Family assistance',exact:true}).click();
  await helperPage.getByRole('button',{name:'Accept helper invitation'}).click();
  await helperPage.getByRole('button',{name:'Find suggestions'}).click();
  const candidate=helperPage.getByRole('article').filter({has:helperPage.getByRole('heading',{name:`@${members[2].handle}`,exact:false})});
  helperPage.once('dialog',dialog=>dialog.accept('Both enjoy hiking'));
  await candidate.getByRole('button',{name:'Suggest this profile'}).click();
  await expect(helperPage.getByText('Suggestion saved. The profile owner decides whether to make contact.')).toBeVisible();
  await page.getByRole('button',{name:'Refresh family assistance'}).click();
  await expect(page.getByText('Both enjoy hiking',{exact:true})).toBeVisible();
  await page.getByRole('button',{name:'Revoke helper'}).click();
  await expect(page.getByText('No current suggestions.')).toBeVisible();
  expect((await helperContext.request.get(base+`/api/matrimony/assistance/${members[0].id}`)).status()).toBe(403);
 }finally{
  for(const r of requests)await r.delete(base+'/api/me',{headers});
  await helperContext.close();await candidateContext.close();
 }
});
