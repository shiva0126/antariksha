import {defineConfig} from '@playwright/test';
export default defineConfig({testDir:'./tests',use:{baseURL:'http://127.0.0.1:3000',headless:true,launchOptions:{executablePath:process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE}},workers:1,timeout:60000});
