import fs from 'node:fs';
import path from 'node:path';
import zlib from 'node:zlib';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

function createPng(width, height, r, g, b, a = 255) {
  // Signature
  const signature = Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]);

  // IHDR chunk
  const ihdrData = Buffer.alloc(13);
  ihdrData.writeUInt32BE(width, 0);
  ihdrData.writeUInt32BE(height, 4);
  ihdrData.writeUInt8(8, 8); // 8-bit depth
  ihdrData.writeUInt8(6, 9); // RGBA color type
  ihdrData.writeUInt8(0, 10); // deflate compression
  ihdrData.writeUInt8(0, 11); // default filter
  ihdrData.writeUInt8(0, 12); // non-interlaced

  const ihdrChunk = createChunk('IHDR', ihdrData);

  // Raw image data with scanline filter 0
  const rowSize = 1 + width * 4;
  const rawData = Buffer.alloc(rowSize * height);

  for (let y = 0; y < height; y++) {
    const rowOffset = y * rowSize;
    rawData[rowOffset] = 0; // Filter None
    for (let x = 0; x < width; x++) {
      const pixelOffset = rowOffset + 1 + x * 4;
      // Draw a subtle border and styled center
      const isBorder = x < 8 || x >= width - 8 || y < 8 || y >= height - 8;
      const isCenter = Math.abs(x - width / 2) < width / 4 && Math.abs(y - height / 2) < height / 4;

      if (isBorder) {
        rawData[pixelOffset] = 217;     // Ochre border R
        rawData[pixelOffset + 1] = 119; // G
        rawData[pixelOffset + 2] = 6;   // B
        rawData[pixelOffset + 3] = 255;
      } else if (isCenter) {
        rawData[pixelOffset] = 251;     // Amber center R
        rawData[pixelOffset + 1] = 191; // G
        rawData[pixelOffset + 2] = 36;  // B
        rawData[pixelOffset + 3] = 255;
      } else {
        rawData[pixelOffset] = r;
        rawData[pixelOffset + 1] = g;
        rawData[pixelOffset + 2] = b;
        rawData[pixelOffset + 3] = a;
      }
    }
  }

  const compressedData = zlib.deflateSync(rawData);
  const idatChunk = createChunk('IDAT', compressedData);
  const iendChunk = createChunk('IEND', Buffer.alloc(0));

  return Buffer.concat([signature, ihdrChunk, idatChunk, iendChunk]);
}

function crc32(buf) {
  let c = 0xffffffff;
  for (let i = 0; i < buf.length; i++) {
    c = (c >>> 8) ^ crcTable[(c ^ buf[i]) & 0xff];
  }
  return (c ^ 0xffffffff) >>> 0;
}

const crcTable = new Uint32Array(256);
for (let n = 0; n < 256; n++) {
  let c = n;
  for (let k = 0; k < 8; k++) {
    c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
  }
  crcTable[n] = c;
}

function createChunk(type, data) {
  const len = data.length;
  const chunk = Buffer.alloc(4 + 4 + len + 4);
  chunk.writeUInt32BE(len, 0);
  chunk.write(type, 4, 4, 'ascii');
  data.copy(chunk, 8);
  const typeAndData = chunk.subarray(4, 8 + len);
  const crcVal = crc32(typeAndData);
  chunk.writeUInt32BE(crcVal, 8 + len);
  return chunk;
}

const targetDir = path.resolve(__dirname, '../public/icons');
fs.mkdirSync(targetDir, { recursive: true });

// Ink background: rgb(24, 24, 27)
fs.writeFileSync(path.join(targetDir, 'icon-192x192.png'), createPng(192, 192, 24, 24, 27));
fs.writeFileSync(path.join(targetDir, 'icon-512x512.png'), createPng(512, 512, 24, 24, 27));
fs.writeFileSync(path.join(targetDir, 'maskable-icon-512x512.png'), createPng(512, 512, 24, 24, 27));
fs.writeFileSync(path.resolve(__dirname, '../public/apple-touch-icon.png'), createPng(180, 180, 24, 24, 27));
fs.writeFileSync(path.resolve(__dirname, '../public/favicon.ico'), createPng(32, 32, 24, 24, 27));

console.log('Icons generated successfully.');
