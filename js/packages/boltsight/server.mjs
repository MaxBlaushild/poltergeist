import { createServer as createHttpServer } from 'node:http';
import { createReadStream, existsSync } from 'node:fs';
import { stat } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const packageDir = path.dirname(fileURLToPath(import.meta.url));
const publicDir = path.join(packageDir, 'public');
const distDir = path.join(packageDir, 'dist');
const defaultApiOrigin = 'https://api.unclaimedstreets.com';

const mimeTypes = new Map([
  ['.css', 'text/css; charset=utf-8'],
  ['.html', 'text/html; charset=utf-8'],
  ['.ico', 'image/x-icon'],
  ['.jpg', 'image/jpeg'],
  ['.jpeg', 'image/jpeg'],
  ['.js', 'text/javascript; charset=utf-8'],
  ['.json', 'application/json; charset=utf-8'],
  ['.png', 'image/png'],
  ['.svg', 'image/svg+xml'],
  ['.webp', 'image/webp']
]);

export function createAppServer(options = {}) {
  const staticRoot = options.staticRoot || getDefaultStaticRoot();
  const apiOrigin = options.apiOrigin || process.env.BOLTSIGHT_API_ORIGIN || defaultApiOrigin;

  return createHttpServer(async (request, response) => {
    try {
      const url = new URL(request.url || '/', 'http://localhost');

      if (url.pathname === '/trades-ar-glasses/interest' || url.pathname === '/api/interest') {
        await proxyInterestRequest(request, response, apiOrigin);
        return;
      }

      if (request.method !== 'GET' && request.method !== 'HEAD') {
        sendJson(response, 405, { error: 'method not allowed' });
        return;
      }

      await serveStatic(request, response, staticRoot, url.pathname);
    } catch (error) {
      const statusCode = Number(error.statusCode) || 500;
      const message = statusCode >= 500 ? 'internal server error' : error.message;
      sendJson(response, statusCode, { error: message });
    }
  });
}

async function proxyInterestRequest(request, response, apiOrigin) {
  const body = await readRequestBody(request);
  const upstream = await fetch(`${apiOrigin}/trades-ar-glasses/interest`, {
    method: request.method,
    headers: {
      'content-type': request.headers['content-type'] || 'application/json',
      accept: request.headers.accept || 'application/json',
      'user-agent': request.headers['user-agent'] || 'boltsight-static-dev-server'
    },
    body: request.method === 'GET' || request.method === 'HEAD' ? undefined : body
  });

  response.writeHead(upstream.status, {
    'content-type': upstream.headers.get('content-type') || 'application/json; charset=utf-8',
    'cache-control': 'no-store'
  });
  response.end(await upstream.text());
}

async function readRequestBody(request) {
  const chunks = [];
  for await (const chunk of request) {
    chunks.push(chunk);
  }
  return Buffer.concat(chunks);
}

async function serveStatic(request, response, staticRoot, pathname) {
  const normalizedPath = pathname === '/' ? '/index.html' : pathname;
  const filePath = path.resolve(staticRoot, decodeURIComponent(normalizedPath).replace(/^\/+/, ''));

  if (!filePath.startsWith(`${staticRoot}${path.sep}`) && filePath !== staticRoot) {
    sendJson(response, 403, { error: 'forbidden' });
    return;
  }

  let fileStat;
  try {
    fileStat = await stat(filePath);
  } catch {
    sendJson(response, 404, { error: 'not found' });
    return;
  }

  if (!fileStat.isFile()) {
    sendJson(response, 404, { error: 'not found' });
    return;
  }

  response.writeHead(200, {
    'content-length': fileStat.size,
    'content-type': mimeTypes.get(path.extname(filePath)) || 'application/octet-stream',
    'cache-control': filePath.endsWith('index.html') ? 'no-cache' : 'public, max-age=3600'
  });

  if (request.method === 'HEAD') {
    response.end();
    return;
  }

  createReadStream(filePath).pipe(response);
}

function getDefaultStaticRoot() {
  if (process.env.NODE_ENV === 'production' && existsSync(distDir)) {
    return distDir;
  }

  return publicDir;
}

function sendJson(response, statusCode, payload) {
  response.writeHead(statusCode, {
    'content-type': 'application/json; charset=utf-8',
    'cache-control': 'no-store'
  });
  response.end(JSON.stringify(payload));
}

const isMain = process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url);
if (isMain) {
  const port = Number(process.env.PORT || 4177);
  const host = process.env.HOST || '127.0.0.1';
  createAppServer().listen(port, host, () => {
    console.log(`BoltSight landing page listening on http://${host}:${port}`);
    console.log(`Interest form submissions proxy to ${process.env.BOLTSIGHT_API_ORIGIN || defaultApiOrigin}`);
  });
}
