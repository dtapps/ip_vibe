#!/usr/bin/env node
'use strict';

// 把同一次发布的附件从**单一平台**搬到 assets/，写出 gitee-release.json 清单，
// 交给 gitee-release-publish.js 建 Release 并上传。
//
// DOWNLOAD_SOURCE=cnb（默认）/ github 决定这一次发布的唯一来源，绝不混用：
// 包里注入了 RepoSource，客户端据此选升级通道；同一个 Release 里混入两个平台的
// 附件，等于按用户下载到哪个包随机分配升级通道。
//   - cnb：附件名是固定的（见 CNB_FILES），URL 按仓库 + tag + 文件名拼出来；
//     大小不能用 GitHub 的 asset.size（两个平台各自编译，同版本字节数不同），
//     改用 Range: bytes=1- 从 Content-Range 读真实总长；正文在本地生成，
//     全程不碰 GitHub 的 API（CNB API 匿名 401，GitHub 元数据一律不读）。
//   - github：清单 / 大小 / 正文都取 GitHub Release 元数据，附件用
//     browser_download_url，下载依次走 daocloud / ghfast / gh-proxy / 直连。
//
// 只用 Node 内置模块：Gitee 的固定构建容器里没有 python3 / jq。
// marked 仅用于 markdown→HTML，缺了会自动退回内置的极简渲染器。
//
// 环境变量：
//   DOWNLOAD_SOURCE       cnb（默认）/ github，由 .workflow/release.yml 归一化后 export
//   VERSION / RELEASE_VERSION  可选，v1.2.3 / 1.2.3 / nightly；留空取来源平台的最新 tag
//   PRERELEASE           可选，由 shell 的 resolve_version 给出；缺省按版本号后缀判定
//   RELEASE_COMMIT       可选，被发布的 commit；Release 的 tag 会指向它
//   OUT_DIR              默认 assets
//   MANIFEST             默认 gitee-release.json
//   CNB_REPO             CNB 仓库路径，默认 dtapp/ip_vibe
//   CNB_BASE             CNB 站点根，默认 https://cnb.cool（本地联调用）
//   GH_REPO              GitHub 仓库路径，默认 dtapps/ip_vibe
//   IPVIBE_GITHUB_TOKEN  可选，避开 api.github.com 匿名限流
//   MAX_FILE_BYTES       默认 100MB（Gitee 单附件上限）
//   MAX_TOTAL_BYTES      默认 800MB（仓库 1GB 留余量）
//
// --resolve-tag：只打印归一化后的最新 tag 就退出（不下载、不写正文）。
//   github 取 GitHub /releases/latest；cnb 用匿名 git ls-remote --tags 取最高版本 tag
//   （CNB API 匿名 401，不能走 API）。shell 在 VERSION 为空时先调它。

const fs = require('fs');
const path = require('path');
const http = require('http');
const https = require('https');
const { execFileSync } = require('child_process');

function fail(msg) {
  console.error('错误：' + msg);
  process.exit(1);
}

function normalizeSource(v) {
  const s = String(v == null ? '' : v).trim().toLowerCase();
  if (!s) return 'cnb';
  if (s !== 'cnb' && s !== 'github') {
    fail('DOWNLOAD_SOURCE 只能是 cnb 或 github，收到：' + JSON.stringify(v));
  }
  return s;
}

const SOURCE = normalizeSource(process.env.DOWNLOAD_SOURCE);
const VERSION_IN = (process.env.VERSION || process.env.RELEASE_VERSION || '').trim();
const OUT_DIR = (process.env.OUT_DIR || 'assets').trim();
const MANIFEST = (process.env.MANIFEST || 'gitee-release.json').trim();
const CNB_REPO = (process.env.CNB_REPO || 'dtapp/ip_vibe').trim();
const CNB_BASE = (process.env.CNB_BASE || 'https://cnb.cool').replace(/\/+$/, '');
const GH_REPO = (process.env.GH_REPO || 'dtapps/ip_vibe').trim();
const GH_TOKEN = (process.env.IPVIBE_GITHUB_TOKEN || process.env.GH_TOKEN || '').trim();
const GH_API = process.env.GH_API || 'https://api.github.com/repos/' + GH_REPO;
const MAX_FILE = Number(process.env.MAX_FILE_BYTES || 100 * 1024 * 1024);
const MAX_TOTAL = Number(process.env.MAX_TOTAL_BYTES || 800 * 1024 * 1024);
const CONCURRENCY = 3;

const VER = VERSION_IN.replace(/^v/i, '');
let TAG = VER ? (/^nightly$/i.test(VER) ? 'nightly' : 'v' + VER) : '';

// 境内直连 GitHub 附件只有 10~30 KB/s 且常被掐断，先走代理镜像，直连只做兜底。
// 镜像里还是 GitHub 同一个文件，字节相同；cnb 源不查这张表（只从 cnb.cool 拿）。
const MIRRORS = [
  { name: 'daocloud', wrap: (u) => 'https://files.m.daocloud.io/' + u.replace(/^https?:\/\//, '') },
  { name: 'ghfast', wrap: (u) => 'https://ghfast.top/' + u },
  { name: 'gh-proxy', wrap: (u) => 'https://gh-proxy.com/' + u },
  { name: '直连', wrap: (u) => u }
];

function mb(n) { return (n / 1048576).toFixed(1) + ' MB'; }
function rate(bps) {
  if (bps >= 1048576) return mb(bps) + '/s';
  if (bps >= 1024) return (bps / 1024).toFixed(0) + ' KB/s';
  return bps.toFixed(0) + ' B/s';
}
function sleep(ms) { return new Promise((r) => setTimeout(r, ms)); }

function prereleaseBySuffix(tag) {
  return /[-.](rc|beta|alpha|nightly)\d*(\.|\b)/i.test(tag || '');
}

// cnb 源：附件名固定（与另外四个平台逐字一致）。可通过 CNB_FILES 覆盖，
// 默认 3 个服务 × 6 个平台 = 18 个压缩包 + SHA256SUMS。
const TARGETS = [
  ['linux', 'amd64', 'tar.gz', 'Linux x86_64'],
  ['linux', 'arm64', 'tar.gz', 'Linux ARM64'],
  ['windows', 'amd64', 'zip', 'Windows x86_64'],
  ['windows', 'arm64', 'zip', 'Windows ARM64'],
  ['darwin', 'amd64', 'tar.gz', 'macOS x86_64'],
  ['darwin', 'arm64', 'tar.gz', 'macOS Apple Silicon']
];
function cnbFiles() {
  if (process.env.CNB_FILES) {
    return process.env.CNB_FILES.split(',').map((s) => s.trim()).filter(Boolean)
      .map((name) => ({ name, desc: '安装包' }));
  }
  const out = [];
  for (const svc of ['backend', 'reporter', 'forwarder']) {
    for (const t of TARGETS) {
      out.push({
        name: svc + '-' + t[0] + '-' + t[1] + '.' + t[2],
        os: t[0] === 'windows' ? 'Windows' : (t[0] === 'darwin' ? 'macOS' : 'Linux'),
        arch: t[1] === 'amd64' ? 'x86_64' : 'ARM64',
        desc: t[3]
      });
    }
  }
  out.push({ name: 'SHA256SUMS', desc: 'SHA256 校验和' });
  return out;
}

function urlFor(name) {
  if (SOURCE === 'cnb') {
    return CNB_BASE + '/' + CNB_REPO + '/-/releases/download/' + TAG + '/' + name;
  }
  return null; // github 源只认 Release 元数据里的 browser_download_url
}

// 断点续传要带 Range 头；重定向到签名地址时也得一路带着，否则续传变成从头下。
function request(url, redirects, withAuth, rangeFrom) {
  return new Promise((resolve, reject) => {
    const lib = url.indexOf('https:') === 0 ? https : http;
    const headers = { 'User-Agent': 'ip-vibe-release-sync' };
    if (GH_TOKEN && withAuth) headers.Authorization = 'token ' + GH_TOKEN;
    if (rangeFrom > 0) headers.Range = 'bytes=' + rangeFrom + '-';
    const req = lib.get(url, { headers }, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        res.resume();
        if (!redirects) return reject(new Error('重定向次数过多: ' + url));
        return resolve(request(new URL(res.headers.location, url).href, redirects - 1, withAuth, rangeFrom));
      }
      if (res.statusCode !== 200 && res.statusCode !== 206) {
        let body = '';
        res.on('data', (c) => { body += c; });
        res.on('end', () => {
          reject(new Error('HTTP ' + res.statusCode + ' ' + url + '\n' + body.slice(0, 300)));
        });
        return;
      }
      resolve(res);
    });
    req.on('error', reject);
    req.setTimeout(60000, () => req.destroy(new Error('请求超时: ' + url)));
  });
}

function readText(url) {
  return request(url, 5, true).then((res) => new Promise((resolve, reject) => {
    let body = '';
    res.setEncoding('utf8');
    res.on('data', (c) => { body += c; });
    res.on('end', () => resolve(body));
    res.on('error', reject);
  }));
}

// CNB API 匿名 401，所以用匿名 git ls-remote 取 tag；不碰 GitHub 的接口。
function latestCnbTag() {
  let out = '';
  try {
    out = execFileSync('git', ['ls-remote', '--tags', '--refs', 'https://cnb.cool/' + CNB_REPO + '.git'], {
      encoding: 'utf8', timeout: 30000, stdio: ['ignore', 'pipe', 'pipe']
    });
  } catch (err) {
    fail('匿名读取 CNB tag 失败（git ls-remote：' +
      (err && err.message ? err.message : String(err)) + '）。请手工设置 VERSION。');
  }
  const tags = out.split('\n').map((l) => {
    const i = l.indexOf('refs/tags/');
    return i === -1 ? '' : l.slice(i + 'refs/tags/'.length).trim();
  }).filter((t) => /^[vV]?[0-9]+\.[0-9]+\.[0-9]+/.test(t));
  if (!tags.length) fail('CNB 仓库没有可用的版本 tag，请手工设置 VERSION');
  tags.sort(compareTag);
  // 与 GitHub /releases/latest 同口径：有正式版就取最新正式版，否则取最高预发布。
  const stable = tags.filter((t) => !/-(rc|beta|alpha|nightly)/i.test(t));
  const picked = stable.length ? stable[stable.length - 1] : tags[tags.length - 1];
  return picked.replace(/^[vV]/, 'v');
}

function compareTag(a, b) {
  const pa = tagNums(a);
  const pb = tagNums(b);
  for (let i = 0; i < 3; i++) {
    if (pa[i] !== pb[i]) return pa[i] - pb[i];
  }
  const qa = /-(rc|beta|alpha|nightly)(\.?\d+)?/i.exec(a);
  const qb = /-(rc|beta|alpha|nightly)(\.?\d+)?/i.exec(b);
  if (!qa) return 1;
  if (!qb) return -1;
  const rank = { alpha: 1, beta: 2, rc: 3, nightly: 4 };
  if (rank[qa[1].toLowerCase()] !== rank[qb[1].toLowerCase()]) {
    return rank[qa[1].toLowerCase()] - rank[qb[1].toLowerCase()];
  }
  return Number((qa[2] || '0').replace(/\D/g, '') || 0) - Number((qb[2] || '0').replace(/\D/g, '') || 0);
}

function tagNums(t) {
  const m = /^[vV]?([0-9]+)\.([0-9]+)\.([0-9]+)/.exec(t);
  return m ? [Number(m[1]), Number(m[2]), Number(m[3])] : [0, 0, 0];
}

async function latestTag() {
  if (SOURCE === 'cnb') return latestCnbTag();
  const rel = JSON.parse(await readText(GH_API + '/releases/latest'));
  const tag = String(rel.tag_name || '').trim();
  if (!tag) fail('GitHub 最新 Release 没有 tag_name，无法判定版本');
  return tag;
}

// cnb 源的大小不能复用 GitHub asset.size：两个平台各自编译，同版本字节数不同。
// 发一个 Range: bytes=1- 的请求，从 Content-Range 尾部读真实总长度。
function probeSize(url) {
  return request(url, 5, false, 1).then((res) => {
    const cr = res.headers['content-range'];
    res.destroy();
    if (cr) {
      const m = /\/(\d+)\s*$/.exec(cr);
      if (m) return Number(m[1]);
    }
    const cl = Number(res.headers['content-length']);
    return res.statusCode === 200 && cl > 0 ? cl : 0;
  });
}

// 落盘后验文件头：反代拿 200 回一个 HTML 错误页时，大小可能对得上，内容却全是文本。
// .deb 必须是 ar 归档（!<arch>\n），.fpk / .tar.gz 必须是 gzip（1f 8b），
// .zip 必须是 PK 头；SHA256SUMS 是文本，至少不能是 HTML 错误页。
function looksRight(dest, name) {
  let fd;
  try {
    fd = fs.openSync(dest, 'r');
    const buf = Buffer.alloc(8);
    const got = fs.readSync(fd, buf, 0, 8, 0);
    if (got < 2) return false;
    if (/\.deb$/i.test(name)) return buf.slice(0, 8).toString('ascii') === '!<arch>\n';
    if (/\.(fpk|tgz)$/i.test(name) || /\.tar\.gz$/i.test(name)) return buf[0] === 0x1f && buf[1] === 0x8b;
    if (/\.zip$/i.test(name)) {
      const sig = buf.slice(0, 4).toString('latin1');
      return sig === 'PK\x03\x04' || sig === 'PK\x05\x06' || sig === 'PK\x07\x08';
    }
    if (/sha256sums/i.test(name)) {
      const head = buf.slice(0, got).toString('utf8').toLowerCase();
      return head.indexOf('<!doctype') === -1 && head.indexOf('<html') === -1;
    }
    return true;
  } catch (e) {
    return false;
  } finally {
    if (fd !== undefined) fs.closeSync(fd);
  }
}

function downloadOnce(url, dest, from) {
  return request(url, 5, false, from).then((res) => new Promise((resolve, reject) => {
    // 只有服务端确认 206 才续写，否则截断重来
    const append = from > 0 && res.statusCode === 206;
    const out = fs.createWriteStream(dest, { flags: append ? 'a' : 'w' });
    let written = append ? from : 0;
    let lastAt = Date.now();
    let lastBytes = written;
    let lowStreak = 0;
    let done = false;

    function abort(err) {
      if (done) return;
      done = true;
      out.destroy();
      res.destroy();
      reject(err);
    }

    res.on('data', (chunk) => {
      written += chunk.length;
      const now = Date.now();
      if (now - lastAt < 5000) return;
      const bps = (written - lastBytes) / ((now - lastAt) / 1000);
      console.log('    ' + mb(written) + '  ' + rate(bps));
      // 15 秒持续低于 30 KB/s 判定这条源废了，主动断开换源：
      // 慢而不断的连接比断掉更耗时间。
      lowStreak = bps < 30 * 1024 ? lowStreak + 1 : 0;
      if (lowStreak >= 3) {
        const err = new Error('速度过低（' + rate(bps) + '）');
        err.slow = true;
        abort(err);
        return;
      }
      lastAt = now;
      lastBytes = written;
    });
    res.on('error', abort);
    res.on('aborted', () => abort(new Error('连接被中断')));
    out.on('error', abort);
    out.on('finish', () => {
      if (done) return;
      done = true;
      resolve(written);
    });
    res.pipe(out);
  }));
}

// 单个文件：累计重试 8 次，每次从已落盘字节续传；慢源立刻断、普通失败两轮换源。
function downloadFile(name, url, dest, expectedSize) {
  const chain = SOURCE === 'cnb' ? [{ name: 'CNB', wrap: (u) => u }] : MIRRORS;
  let idx = 0;
  let tries = 0;

  function step() {
    const from = fs.existsSync(dest) ? fs.statSync(dest).size : 0;
    if (expectedSize > 0 && from === expectedSize) {
      console.log('  已完整（' + mb(from) + '）');
      return Promise.resolve();
    }
    if (expectedSize > 0 && from > expectedSize) fs.unlinkSync(dest);
    return downloadOnce(chain[idx].wrap(url), dest, from).then((written) => {
      if (expectedSize > 0 && written !== expectedSize) {
        throw new Error('大小不符：' + written + ' != ' + expectedSize);
      }
      if (!looksRight(dest, name)) {
        // 错误页删掉再重试，否则下次 from === expectedSize 会被误判为「已完整」
        fs.unlinkSync(dest);
        throw new Error('文件头不对，可能是反代返回的 HTML 错误页');
      }
      console.log('  完成 ' + mb(written) + '（' + chain[idx].name + '）');
    }).catch((err) => {
      tries++;
      const now = fs.existsSync(dest) ? fs.statSync(dest).size : 0;
      console.log('  中断：' + (err && err.message ? err.message : err) +
        '，已下 ' + mb(now) + '，第 ' + tries + ' 次');
      if (tries >= 8) throw new Error('下载失败（重试 ' + tries + ' 次）：' + url);
      if ((err && err.slow) || tries % 2 === 0) {
        if (idx < chain.length - 1) {
          idx++;
          console.log('  切换下载源 → ' + chain[idx].name);
        }
      }
      return sleep(1500).then(step);
    });
  }
  return step();
}

// 把某一整节（"## 标题" 到下一个 "## "）的内容换成 replacement，标题保留。
function replaceSection(text, title, replacement) {
  const out = [];
  let inside = false;
  for (const line of text.split('\n')) {
    if (/^##\s+/.test(line)) {
      inside = line.trim() === '## ' + title;
      out.push(line);
      if (inside) { out.push(''); out.push(replacement); }
      continue;
    }
    if (!inside) out.push(line);
  }
  return out.join('\n');
}

function escapeHtml(s) {
  return String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

// marked 不在时也能出可用的 HTML：本项目的正文只用到标题 / 表格 / 列表 / 引用 /
// 链接 / 加粗 / 行内代码，这几个结构足够；外部 release body 的复杂写法会退化为文本。
function minimalMarkdown(md) {
  const lines = String(md).replace(/\r\n/g, '\n').split('\n');
  const out = [];
  let inUl = false;
  let inTable = false;
  function inline(s) {
    return escapeHtml(s)
      .replace(/`([^`]+)`/g, '<code>$1</code>')
      .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
      .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2">$1</a>');
  }
  for (const raw of lines) {
    const line = raw.trim();
    if (/^\|/.test(line)) {
      if (!inTable) { out.push('<table>'); inTable = true; }
      const cells = line.replace(/^\||\|$/g, '').split('|').map((c) => c.trim());
      if (cells.every((c) => /^:?-{3,}:?$/.test(c))) continue;
      out.push('<tr>' + cells.map((c) => '<td>' + inline(c) + '</td>').join('') + '</tr>');
      continue;
    }
    if (inTable) { out.push('</table>'); inTable = false; }
    if (/^#{1,6}\s+/.test(line)) {
      const m = /^(#{1,6})\s+(.*)$/.exec(line);
      out.push('<h' + m[1].length + '>' + inline(m[2]) + '</h' + m[1].length + '>');
    } else if (/^>\s?/.test(line)) {
      out.push('<blockquote>' + inline(line.replace(/^>\s?/, '')) + '</blockquote>');
    } else if (/^[-*]\s+/.test(line)) {
      if (!inUl) { out.push('<ul>'); inUl = true; }
      out.push('<li>' + inline(line.replace(/^[-*]\s+/, '')) + '</li>');
    } else {
      if (inUl) { out.push('</ul>'); inUl = false; }
      if (line) out.push('<p>' + inline(line) + '</p>');
    }
  }
  if (inUl) out.push('</ul>');
  if (inTable) out.push('</table>');
  return out.join('\n');
}

// GitHub markdown → Gitee 正文 HTML。保留：预发布提示 / 发布信息 / 下载 / 校验 /
// 更新内容（标题）；「更新内容」明细换成指引，完整清单看仓库的 CHANGELOG。
function renderBody(md) {
  let html;
  try {
    const imported = require('marked');
    const marked = (imported && imported.marked) ? imported.marked : imported;
    html = (typeof marked.parse === 'function' ? marked.parse : marked)(
      replaceSection(md, '更新内容', '> 完整更新内容请查看仓库的 CHANGELOG.md。')
    );
  } catch (e) {
    html = minimalMarkdown(replaceSection(md, '更新内容', '> 完整更新内容请查看仓库的 CHANGELOG.md。'));
  }
  // Gitee 页面自带附件列表，正文里的链接点了也是跳走，只留链接文字。
  html = html.replace(
    /<a\b[^>]*\bhref="([^"]*)"[^>]*>([\s\S]*?)<\/a>/g,
    (m, url, text) => text.replace(/<[^>]+>/g, '').trim()
  ).replace(/<a\b[^>]*>\s*<\/a>/g, '');
  console.log('正文：' + Buffer.byteLength(html) + ' 字节 HTML');
  return html;
}

// cnb 源本地生成的兜底正文：工作流随后会用五平台共用的 render-release-body.sh
// 覆盖 release_body.md；只有那个脚本在 Gitee 环境跑不起来时才会用到这里。
function localBody(sizes) {
  const pre = (process.env.PRERELEASE || '').trim() === 'true' || prereleaseBySuffix(TAG);
  const dl = 'https://gitee.com/dtapps/ip_vibe/releases/download/' + TAG;
  const now = new Date().toISOString().replace('T', ' ').slice(0, 19) + ' UTC';
  const lines = [];
  if (pre) {
    lines.push('## ⚠️ 预发布版本', '', '> 这是预发布版本，可能包含未充分测试的功能。', '');
  }
  lines.push('## 发布信息', '', '| 项目 | 内容 |', '| --- | --- |');
  lines.push('| 版本号 | ' + TAG + ' |');
  lines.push('| 发布类型 | ' + (pre ? '🚧 预发布版本 (Pre-release)' : '✅ 正式发布版本 (Stable)') + ' |');
  lines.push('| 构建时间 | ' + now + ' |');
  lines.push('', '## 下载', '', '| 平台 | 架构 | 文件 | 大小 | 说明 |', '| --- | --- | --- | --- | --- |');
  for (const f of cnbFiles()) {
    lines.push('| ' + (f.os || 'Linux') + ' | ' + (f.arch || 'x86_64') + ' | [' + f.name +
      '](' + dl + '/' + f.name + ') | ' + mb(sizes[f.name] || 0) + ' | ' + f.desc + ' |');
  }
  lines.push('', '## 校验', '', '把 `SHA256SUMS` 和压缩包下载到同一目录，然后按平台执行：', '',
    '- **Linux**：`sha256sum -c SHA256SUMS`', '- **macOS**：`shasum -a 256 -c SHA256SUMS`',
    '', '## 更新内容', '', '- 详见 [CHANGELOG.md](https://gitee.com/dtapps/ip_vibe/blob/' + TAG + '/CHANGELOG.md)', '');
  return lines.join('\n');
}

async function runCnb() {
  console.log('附件来源：CNB（固定，不混用；正文也在本地生成，不查 GitHub）');
  console.log('Release ' + TAG + '：' + cnbFiles().length + ' 个固定产物');

  const sizes = {};
  let seq = 0;
  const queue = cnbFiles().slice();

  async function downloadOne(item) {
    const n = ++seq;
    const dest = path.join(OUT_DIR, item.name);
    const url = urlFor(item.name);
    const expected = await probeSize(url).catch(() => 0);
    if (expected > MAX_FILE) {
      console.log('[' + n + '/' + cnbFiles().length + '] ' + item.name +
        '（' + mb(expected) + ' 超过 Gitee 单附件 ' + mb(MAX_FILE) + '，跳过该文件）');
      return;
    }
    console.log('[' + n + '/' + cnbFiles().length + '] ' + item.name + '（' + mb(expected) + '）');
    await downloadFile(item.name, url, dest, expected);
    sizes[item.name] = fs.statSync(dest).size;
  }

  async function worker() {
    while (queue.length) await downloadOne(queue.shift());
  }

  const pool = [];
  for (let i = 0; i < Math.min(CONCURRENCY, queue.length); i++) pool.push(worker());
  await Promise.all(pool);

  const files = Object.keys(sizes).sort().map((name) => ({
    name, size: sizes[name], path: path.join(OUT_DIR, name)
  }));
  return { files, body: localBody(sizes), targetCommitish: 'master' };
}

async function runGithub() {
  const relUrl = TAG
    ? GH_API + '/releases/tags/' + encodeURIComponent(TAG)
    : GH_API + '/releases/latest';
  console.log('版本号: ' + (TAG || '未指定，取 GitHub 最新 Release'));
  const rel = JSON.parse(await readText(relUrl));
  if (!TAG) {
    TAG = String(rel.tag_name || '').trim();
    if (!TAG) fail('GitHub 最新 Release 没有 tag_name，无法判定版本');
    console.log('自动判定版本：' + TAG);
  }

  const all = rel.assets || [];
  if (!all.length) fail('GitHub Release ' + TAG + ' 没有附件');
  const skipped = [];
  const assets = all.filter((a) => {
    if ((a.size || 0) > MAX_FILE) {
      skipped.push(a.name + '（' + mb(a.size || 0) + ' 超过 Gitee 单附件上限，跳过该文件）');
      return false;
    }
    return true;
  });
  if (!assets.length) fail('过滤后没有可上传 Gitee 的附件');
  const totalBytes = assets.reduce((s, a) => s + (a.size || 0), 0);
  console.log('附件来源：GitHub（固定，不混用，四个镜像依次回退）');
  console.log('共 ' + all.length + ' 个附件，需要下载并上传 ' + assets.length + ' 个（' + mb(totalBytes) + '）');
  skipped.forEach((s) => console.log('  跳过 ' + s));
  if (totalBytes > MAX_TOTAL) {
    fail('待上传 ' + mb(totalBytes) + ' 超过安全阈值 ' + mb(MAX_TOTAL) + '，请先清理旧 Release 的附件');
  }

  let seq = 0;
  const queue = assets.slice();
  async function downloadOne(asset) {
    const n = ++seq;
    const dest = path.join(OUT_DIR, asset.name);
    console.log('[' + n + '/' + assets.length + '] ' + asset.name + '（' + mb(asset.size || 0) + '）');
    await downloadFile(asset.name, asset.browser_download_url, dest, asset.size || 0);
  }
  async function worker() {
    while (queue.length) await downloadOne(queue.shift());
  }
  const pool = [];
  for (let i = 0; i < Math.min(CONCURRENCY, queue.length); i++) pool.push(worker());
  await Promise.all(pool);

  const body = String(rel.body || '').replace(/\r\n/g, '\n')
    .split('https://github.com/dtapps/ip_vibe').join('https://gitee.com/dtapps/ip_vibe').trim();
  return {
    files: assets.map((a) => ({ name: a.name, size: a.size || 0, path: path.join(OUT_DIR, a.name) })),
    body: body || ('IPVibe ' + TAG),
    targetCommitish: String(rel.target_commitish || 'master'),
    skipped
  };
}

function selfTest() {
  const assert = require('assert');
  const os = require('os');
  const checks = [
    ['默认来源是 cnb', () => assert.strictEqual(normalizeSource(''), 'cnb')],
    ['github 来源原样保留', () => assert.strictEqual(normalizeSource(' GitHub '), 'github')],
    ['rc 判预发布', () => assert.strictEqual(prereleaseBySuffix('v1.2.3-rc.1'), true)],
    ['nightly 判预发布', () => assert.strictEqual(prereleaseBySuffix('v1.2.4-nightly.20261116'), true)],
    ['正式版不误判', () => assert.strictEqual(prereleaseBySuffix('v1.2.3-hotfix'), false)],
    ['固定附件 18 包 + SHA256SUMS', () => assert.strictEqual(cnbFiles().length, 19)]
  ];
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'ipvibe-dl-'));
  try {
    const deb = path.join(dir, 'a.deb');
    const fpk = path.join(dir, 'a.fpk');
    const html = path.join(dir, 'a.zip');
    fs.writeFileSync(deb, '!<arch>\n0123456789');
    fs.writeFileSync(fpk, Buffer.from([0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00]));
    fs.writeFileSync(html, '<!DOCTYPE html><html>error</html>');
    checks.push(['.deb 文件头', () => assert.strictEqual(looksRight(deb, 'a.deb'), true)]);
    checks.push(['.fpk 文件头', () => assert.strictEqual(looksRight(fpk, 'a.fpk'), true)]);
    checks.push(['HTML 错误页被拒', () => assert.strictEqual(looksRight(html, 'a.zip'), false)]);
    for (const [name, fn] of checks) {
      fn();
      console.log('  ✓ ' + name);
    }
  } finally {
    fs.rmSync(dir, { recursive: true, force: true });
  }
  console.log('自检通过');
}

async function main() {
  try { process.cwd(); } catch (e) { process.chdir('/'); }
  if (process.argv.indexOf('--self-test') !== -1) { selfTest(); return; }

  // 只解析版本用：打印后退出，不下载、不写正文。
  if (process.argv.indexOf('--resolve-tag') !== -1) {
    const resolved = VERSION_IN ? TAG : await latestTag();
    process.stdout.write(resolved + '\n');
    return;
  }

  if (!TAG) TAG = await latestTag();
  console.log('版本号: ' + TAG + '  DOWNLOAD_SOURCE=' + SOURCE);

  if (!fs.existsSync(OUT_DIR)) fs.mkdirSync(OUT_DIR, { recursive: true });
  const result = SOURCE === 'cnb' ? await runCnb() : await runGithub();

  // 只按版本号后缀判预发布（rc / beta / alpha / nightly），与 resolve-prerelease.sh 同口径；
  // shell 已把 resolve_version 的结果传进来时以前者为准，绝不读另一平台的 prerelease 字段。
  const envPre = (process.env.PRERELEASE || '').trim();
  const pre = envPre === 'true' ? true : (envPre === 'false' ? false : prereleaseBySuffix(TAG));
  const bodyHtml = renderBody(result.body || ('IPVibe ' + TAG));

  const releaseCommit = (process.env.RELEASE_COMMIT || '').trim();
  const totalBytes = result.files.reduce((s, f) => s + (f.size || 0), 0);
  const manifest = {
    tag: TAG,
    ver: TAG.replace(/^v/i, ''),
    prerelease: pre,
    releaseName: 'IPVibe ' + TAG + (pre ? ' (预发布)' : ''),
    targetCommitish: result.targetCommitish || 'master',
    commit: releaseCommit,
    bodyHtml,
    outDir: OUT_DIR,
    totalBytes,
    downloadSource: SOURCE,
    files: result.files,
    skipped: result.skipped || []
  };
  fs.writeFileSync(MANIFEST, JSON.stringify(manifest, null, 2) + '\n', 'utf8');

  console.log('---- 下载完成 ----');
  console.log('清单：' + MANIFEST + '（' + manifest.files.length + ' 个文件，' + mb(totalBytes) + '，来源 ' + SOURCE + '）');
  manifest.files.forEach((f) => console.log('  - ' + f.name + ' ' + mb(f.size)));
}

main().catch((err) => fail(err && err.message ? err.message : String(err)));
