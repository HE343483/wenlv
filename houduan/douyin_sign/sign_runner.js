/**
 * 抖音 a_bogus 签名入口(供 Go 后端以 node 子进程调用)
 *
 * 用法: echo '<JSON>' | node sign_runner.js
 * 输入 JSON: { "mode": "sign", "url": "/aweme/v1/web/general/search/single/?device_platform=webapp&...",
 *             "userAgent": "Mozilla/5.0 ..." }
 * 输出 JSON: { "a_bogus": "...", "error": "" }
 *
 * ⚠️ 本文件只是「桥接层」,真正生成 a_bogus 的混淆 JS 需自行放置到本目录下的 douyin.js。
 *    签名 JS 需按以下任一方式导出签名函数(按顺序探测):
 *      1) module.exports = function (urlWithQuery, userAgent) { return aBogus }
 *      2) module.exports.a_bogus / module.exports.get_a_bogus
 *      3) globalThis.a_bogus / globalThis.get_a_bogus / globalThis.sign
 *    其中 urlWithQuery 为「路径 + 查询串」(如 /aweme/v1/web/general/search/single?aid=6383&...),
 *    与社区通用 douyin.js 的调用约定一致。
 */
'use strict';

const fs = require('fs');
const path = require('path');

const SIGN_DIR = __dirname;
const SIGN_FILE = path.join(SIGN_DIR, 'douyin.js');

const DEFAULT_UA =
  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36';

function readStdin() {
  try {
    return fs.readFileSync(0, 'utf8');
  } catch (e) {
    return '';
  }
}

/** 部分 a_bogus 脚本依赖浏览器环境,这里补最小可用桩,避免直接抛 ReferenceError。 */
function installBrowserShim() {
  if (typeof globalThis.window === 'undefined') {
    globalThis.window = globalThis;
  }
  if (typeof globalThis.self === 'undefined') {
    globalThis.self = globalThis;
  }
  if (typeof globalThis.document === 'undefined') {
    globalThis.document = {
      createElement: () => ({ getContext: () => null, style: {} }),
      addEventListener: () => {},
      removeEventListener: () => {},
      documentElement: { style: {} },
      cookie: '',
    };
  }
  if (typeof globalThis.navigator === 'undefined') {
    globalThis.navigator = { userAgent: DEFAULT_UA, platform: 'Win32', language: 'zh-CN' };
  }
  if (typeof globalThis.screen === 'undefined') {
    globalThis.screen = { width: 1920, height: 1080, availWidth: 1920, availHeight: 1040 };
  }
  if (typeof globalThis.location === 'undefined') {
    globalThis.location = { href: 'https://www.douyin.com/', origin: 'https://www.douyin.com' };
  }
  if (typeof globalThis.requestAnimationFrame === 'undefined') {
    globalThis.requestAnimationFrame = (fn) => setTimeout(() => fn(Date.now()), 0);
  }
}

/** 从候选导出中挑出可用的签名函数。 */
function resolveSignFn() {
  const candidates = [];
  try {
    const mod = require(SIGN_FILE);
    candidates.push(mod, mod && mod.default, mod && mod.a_bogus, mod && mod.get_a_bogus);
  } catch (e) {
    if (e && e.code !== 'MODULE_NOT_FOUND') {
      throw new Error('加载 douyin.js 失败: ' + e.message);
    }
  }
  candidates.push(globalThis.a_bogus, globalThis.get_a_bogus, globalThis.sign);
  for (const c of candidates) {
    if (typeof c === 'function') return c;
  }
  return null;
}

function normalizeResult(value) {
  if (typeof value === 'string') return value;
  if (value && typeof value === 'object') {
    // 兼容返回 { a_bogus: '...' } / { aBogus: '...' } 的脚本
    return value.a_bogus || value.aBogus || value.bogus || '';
  }
  return '';
}

function main() {
  let input = {};
  try {
    const raw = readStdin();
    input = raw ? JSON.parse(raw) : {};
  } catch (e) {
    process.stdout.write(JSON.stringify({ error: '参数解析失败: ' + e.message }));
    return;
  }

  try {
    installBrowserShim();
    if (!fs.existsSync(SIGN_FILE)) {
      process.stdout.write(
        JSON.stringify({ error: '未找到抖音签名脚本 douyin_sign/douyin.js,请先放置后再使用抖音数据源' })
      );
      return;
    }
    const signFn = resolveSignFn();
    if (!signFn) {
      process.stdout.write(
        JSON.stringify({ error: 'douyin.js 未导出可用的签名函数(a_bogus / get_a_bogus / sign)' })
      );
      return;
    }
    const userAgent = input.userAgent || DEFAULT_UA;
    const aBogus = normalizeResult(signFn(input.url || '', userAgent));
    if (!aBogus) {
      process.stdout.write(JSON.stringify({ error: '抖音签名结果为空' }));
      return;
    }
    process.stdout.write(JSON.stringify({ a_bogus: aBogus }));
  } catch (e) {
    process.stdout.write(JSON.stringify({ error: (e && e.message) || String(e) }));
  }
}

main();
