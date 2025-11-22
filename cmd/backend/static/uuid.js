// uuid.js - 支持 v4 和 v7
(function (global) {
    // 安全随机字节数组（优先使用 crypto.getRandomValues）
    function getRandomBytes(len) {
        const bytes = new Uint8Array(len);
        if (typeof crypto !== 'undefined' && crypto.getRandomValues) {
            crypto.getRandomValues(bytes);
        } else if (typeof msCrypto !== 'undefined' && msCrypto.getRandomValues) {
            msCrypto.getRandomValues(bytes); // IE11
        } else {
            // 非安全 fallback（仅用于极端兼容）
            for (let i = 0; i < len; i++) {
                bytes[i] = Math.floor(Math.random() * 256);
            }
        }
        return bytes;
    }

    // UUID v4：完全随机
    function uuidv4() {
        const bytes = getRandomBytes(16);
        bytes[6] = (bytes[6] & 0x0f) | 0x40; // version 4
        bytes[8] = (bytes[8] & 0x3f) | 0x80; // variant RFC4122
        const hex = Array.prototype.map.call(bytes, b => b.toString(16).padStart(2, '0')).join('');
        return [
            hex.slice(0, 8),
            hex.slice(8, 12),
            hex.slice(12, 16),
            hex.slice(16, 20),
            hex.slice(20)
        ].join('-');
    }

    // UUID v7：时间有序（RFC 9562）
    let lastTimestamp = 0;
    let counter = 0;

    function uuidv7() {
        const ts = Date.now(); // 毫秒时间戳
        if (ts > lastTimestamp) {
            lastTimestamp = ts;
            counter = 0;
        } else {
            counter = (counter + 1) & 0x3fff; // 14-bit counter
        }

        // 构建 16 字节 UUID
        const bytes = new Uint8Array(16);

        // 前 6 字节：时间戳（48 位）
        bytes[0] = (ts >>> 40) & 0xff;
        bytes[1] = (ts >>> 32) & 0xff;
        bytes[2] = (ts >>> 24) & 0xff;
        bytes[3] = (ts >>> 16) & 0xff;
        bytes[4] = (ts >>> 8) & 0xff;
        bytes[5] = ts & 0xff;

        // 第7字节：4xxx -> version 7
        bytes[6] = 0x70 | (getRandomBytes(1)[0] & 0x0f);

        // 第8字节：10xx -> variant
        bytes[7] = 0x80 | (getRandomBytes(1)[0] & 0x3f);

        // 第9-10字节：counter（14 bits）
        bytes[8] = (counter >>> 8) & 0xff;
        bytes[9] = counter & 0xff;

        // 最后6字节：随机
        const rand = getRandomBytes(6);
        bytes.set(rand, 10);

        const hex = Array.prototype.map.call(bytes, b => b.toString(16).padStart(2, '0')).join('');
        return [
            hex.slice(0, 8),
            hex.slice(8, 12),
            hex.slice(12, 16),
            hex.slice(16, 20),
            hex.slice(20)
        ].join('-');
    }

    // 导出
    const uuid = {
        v4: uuidv4,
        v7: uuidv7
    };

    if (typeof module !== 'undefined' && module.exports) {
        module.exports = uuid;
    } else if (typeof global !== 'undefined') {
        global.uuid = uuid;
    } else if (typeof window !== 'undefined') {
        window.uuid = uuid;
    }
})(typeof globalThis !== 'undefined' ? globalThis : window || global || this);