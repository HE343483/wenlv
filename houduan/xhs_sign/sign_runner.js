/**
 * 小红书签名入口(供 Go 后端以 node 子进程调用)
 *
 * 用法: echo '<JSON>' | node sign_runner.js
 * 输入 JSON: { "mode": "sign" | "traceid", "api": "/api/sns/web/v1/search/notes",
 *             "data": "<请求体 JSON 字符串>", "a1": "<a1 cookie>", "method": "POST" }
 * 输出 JSON: { "xs": "...", "xt": 1699999999999, "xs_common": "...", "traceId": "..." }
 */
'use strict';

const fs = require('fs');
const path = require('path');
const vm = require('vm');

const SIGN_DIR = __dirname;

function readStdin() {
  try {
    return fs.readFileSync(0, 'utf8');
  } catch (e) {
    return '';
  }
}

function loadXrayTraceId() {
  const pack1 = path.join(SIGN_DIR, 'xhs_xray_pack1.js');
  const pack2 = path.join(SIGN_DIR, 'xhs_xray_pack2.js');
  let source = fs.readFileSync(path.join(SIGN_DIR, 'xhs_xray.js'), 'utf8');
  // 将相对 require 替换为绝对路径,避免从 stdin 以外的目录执行时找不到依赖包
  source = source.replace(/require\((['"])[^'"]*xhs_xray_pack1\.js\1\)/g, () => `require(${JSON.stringify(pack1)})`);
  source = source.replace(/require\((['"])[^'"]*xhs_xray_pack2\.js\1\)/g, () => `require(${JSON.stringify(pack2)})`);
  // vm.runInThisContext 环境没有 require,这里以函数包装方式注入 CommonJS 变量
  const script = `(function (require, module, exports, __dirname, __filename) {\n${source}\n})`;
  const fn = vm.runInThisContext(script, { filename: path.join(SIGN_DIR, 'xhs_xray.js') });
  fn(require, module, module.exports, SIGN_DIR, path.join(SIGN_DIR, 'xhs_xray.js'));
  const traceIdFn = typeof traceId === 'function' ? traceId : globalThis.traceId;
  if (typeof traceIdFn !== 'function') {
    throw new Error('traceId 函数未定义');
  }
  return traceIdFn();
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
    if (input.mode === 'traceid') {
      process.stdout.write(JSON.stringify({ traceId: loadXrayTraceId() }));
      return;
    }

    const xsModule = require(path.join(SIGN_DIR, 'xhs_xs_xsc_56.js'));
    const result = xsModule.get_request_headers_params(
      input.api || '',
      input.data || '',
      input.a1 || '',
      input.method || 'POST'
    );
    process.stdout.write(
      JSON.stringify({
        xs: result.xs,
        xt: result.xt,
        xs_common: result.xs_common,
      })
    );
  } catch (e) {
    process.stdout.write(JSON.stringify({ error: (e && e.message) || String(e) }));
  }
}

main();