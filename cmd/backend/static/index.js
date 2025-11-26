// 滚动超过 300 px 出现
window.addEventListener('scroll', () => {
    document.querySelector('.back-top').classList.toggle('visible', window.scrollY > 300);
});

// API请求封装
const API = {
    get: (url) => fetch(url).then(r => r.json()),
    post: (url, data) => fetch(url, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) }).then(r => r.json()),
    patch: (url, data) => fetch(url, { method: 'PATCH', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) }).then(r => r.json()),
    delete: (url) => fetch(url, { method: 'DELETE' }).then(r => r.json())
}

// 统一给 select 插「请选择」占位
function addPlaceholder(select, text = '请选择') {
    const ph = document.createElement('option');
    ph.value = '';
    ph.text = text;
    ph.textContent = text;
    ph.disabled = true;
    ph.selected = true;
    select.insertBefore(ph, select.firstChild);
}

// 辅助：ISO → MM-DD HH:mm
const fmtDate = (iso) => {
    const d = new Date(iso);
    const MM = String(d.getMonth() + 1).padStart(2, '0');
    const DD = String(d.getDate()).padStart(2, '0');
    const HH = String(d.getHours()).padStart(2, '0');
    const mm = String(d.getMinutes()).padStart(2, '0');
    return `${MM}-${DD} ${HH}:${mm}`;
};

// 打开配置模态框
function openModal(id, titleID, titleName) {
    const modal = document.getElementById(id);
    const title = document.getElementById(titleID);
    if (title) title.textContent = titleName;
    if (modal) modal.classList.add('show');
}

// 关闭模态框
function closeModal(id) {
    const modal = document.getElementById(id);
    if (!modal) return;

    // 1. 隐藏弹窗
    modal.classList.remove('show');

    // 2. 如果有 form 才 reset
    const form = modal.querySelector('form');
    if (form) form.reset();

    // 3. 如果有平台字段才隐藏
    modal.querySelectorAll('.platform-fields').forEach(el => el.classList.add('hidden'));
}

// 一键复制
function copyData(id) {
    const pre = document.getElementById(id);
    navigator.clipboard.writeText(pre.textContent)
        .then(() => {
            const btn = document.querySelector('.ip-popup-copy');
            const old = btn.innerHTML;
            btn.innerHTML = '<svg class="ip-i"><use href="/static/icons.svg#ip-check"/></use></svg> 已复制';
            btn.style.color = 'var(--success)';
            setTimeout(() => { btn.innerHTML = old; btn.style.color = ''; }, 1500);
        })
        .catch(() => alert('复制失败，请手动选择复制'));
}

// --- 平台配置 --- 

// 请选择服务能力 capability
function changeConfigCapability(sel) {
    const opt = sel.options[sel.selectedIndex];
    // console.log("changeConfigCapability", opt);
}

// 请选择平台 platform
function changeConfigPlatform(platform) {
    // 隐藏所有平台字段
    document.querySelectorAll('.platform-fields').forEach(el => el.classList.add('hidden'));

    // 显示选中平台的字段
    const fields = document.getElementById(platform + 'Fields');
    if (fields) {
        fields.classList.remove('hidden');
    }
}

// 组装JSON配置
function buildConfigPlatform() {
    const platform = document.getElementById('config_platform').value;
    let config = {};

    switch (platform) {
        case 'edgeone':
            config = {
                secret_id: document.getElementById('edgeone_secret_id').value,
                secret_key: document.getElementById('edgeone_secret_key').value,
                zone_id: document.getElementById('edgeone_zone_id').value,
            };
            break;

        case 'cloudflare':
            config = {
                api_key: document.getElementById('cloudflare_api_key').value,
                zone_id: document.getElementById('cloudflare_zone_id').value,
            };
            break;

        case 'dnspod':
            config = {
                secret_id: document.getElementById('dnspod_secret_id').value,
                secret_key: document.getElementById('dnspod_secret_key').value,
            };
            break;
    }

    // 设置隐藏字段的值
    document.getElementById('config_json').value = JSON.stringify(config);
}

// 平台配置管理（模态框）
async function openConfigModal(id) {
    document.getElementById('configForm').reset();

    // 编辑
    if (id) {
        const { data: config } = await API.get(`/api/configs/${id}`);
        console.log("config", config);
        if (!config) {
            alert('配置数据异常');
            return;
        }

        document.getElementById('config_id').value = config.id;
        document.getElementById('config_id').disabled = true;
        document.getElementById('config_capability').value = config.capability;
        document.getElementById('config_name').value = config.name;
        document.getElementById('config_platform').value = config.platform;
        document.getElementById('config_platform').disabled = true;

        // 解析JSON并填充表单
        changeConfigPlatform(config.platform);

        switch (config.platform) {
            case 'edgeone':
                document.getElementById('edgeone_secret_id').value = config.config_json.secret_id || '';
                document.getElementById('edgeone_secret_key').value = config.config_json.secret_key || '';
                document.getElementById('edgeone_zone_id').value = config.config_json.zone_id || '';
                break;
            case 'cloudflare':
                document.getElementById('cloudflare_api_key').value = config.config_json.api_key || '';
                document.getElementById('cloudflare_zone_id').value = config.config_json.zone_id || '';
                break;
            case 'dnspod':
                document.getElementById('dnspod_secret_id').value = config.config_json.secret_id || '';
                document.getElementById('dnspod_secret_key').value = config.config_json.secret_key || '';
                break;
        }

        openModal('configModal', 'configModalTitle', '编辑 平台配置');
    }

    // 创建
    if (!id) {
        document.getElementById('config_id').value = '';
        document.getElementById('config_id').disabled = true;
        document.querySelectorAll('.platform-fields').forEach(el => el.classList.add('hidden'));
        document.getElementById('config_platform').disabled = false;

        openModal('configModal', 'configModalTitle', '添加 平台配置');
    }

}

// 平台配置管理（模态框） -> 表单提交
document.getElementById('configForm').onsubmit = async (e) => {
    e.preventDefault();
    buildConfigPlatform(); // 组装JSON

    // 表单数据
    const form_id = Number(document.getElementById('config_id').value.trim());
    const form_data = {
        capability: document.getElementById('config_capability').value.trim(),
        name: document.getElementById('config_name').value.trim(),
        platform: document.getElementById('config_platform').value.trim(),
        config_json: document.getElementById('config_json').value.trim(),
    };
    console.log("form_id, form_data", form_id, form_data);

    // 校验
    if (!form_data.capability) {
        return alert('请选择服务能力')
    };
    if (!form_data.name) {
        return alert('请输入组名')
    };
    if (!form_data.platform) {
        return alert('请选择平台')
    };

    // 提交
    if (isNaN(form_id) || form_id <= 0) {
        await API.post('/api/configs', form_data);
    } else {
        await API.patch(`/api/configs/${form_id}`, form_data);
    }

    // 刷新页面
    location.reload();
};

// 删除
async function deleteConfig(id) {
    if (confirm('确定删除此配置？相关记录也会被删除')) {
        await API.delete(`/api/configs/${id}`);

        // 刷新页面
        location.reload();
    }
}

// --- 平台配置 --- 
// --- CDN记录管理 --- 

// 状态
async function toggleCdnRecordStatus(id, checkbox) {
    await API.patch(`/api/cdn-records/${id}/active`, { is_active: checkbox.checked });
}

// CDN记录管理
async function openCdnRecordModal(id) {

    // IP 组列表
    const { data: ipGroups } = await API.get(`/api/ip-groups`);
    console.log("ipGroups", ipGroups);
    if (ipGroups) {
        const ipGroupSelect = document.getElementById('cdn_record_ip_group_id');
        ipGroupSelect.innerHTML = '';
        ipGroups.forEach(d => {
            const opt = document.createElement('option');
            opt.value = d.id;
            opt.text = `${d.name}`;
            ipGroupSelect.appendChild(opt);
        });
    }

    // 编辑
    if (id) {
        const { data: cdnRecord } = await API.get(`/api/cdn-records/${id}`);
        console.log("cdnRecord", cdnRecord);
        if (!cdnRecord) {
            alert('CDN记录数据异常');
            return;
        }

        document.getElementById('cdn_record_id').value = cdnRecord.id;
        document.getElementById('cdn_record_id').disabled = true;
        document.getElementById('cdn_record_config_id').value = cdnRecord.config_id;
        document.getElementById('cdn_record_config_id').disabled = true;
        document.getElementById('cdn_record_domain').value = cdnRecord.domain;
        document.getElementById('cdn_record_domain').disabled = true;
        document.getElementById('cdn_record_type').value = cdnRecord.type;
        document.getElementById('cdn_record_type').disabled = true;
        document.getElementById('cdn_record_origin_protocol').value = cdnRecord.origin_protocol;
        document.getElementById('cdn_record_http_port').value = cdnRecord.http_port;
        document.getElementById('cdn_record_https_port').value = cdnRecord.https_port;
        document.getElementById('cdn_record_ip_group_id').value = cdnRecord.ip_group_id;

        openModal('cdnRecordModal', 'cdnRecordModalTitle', '编辑 CDN记录');
    }

    // 创建
    if (!id) {
        document.getElementById('cdn_record_id').value = 0;
        document.getElementById('cdn_record_id').disabled = true;
        document.getElementById('cdn_record_config_id').disabled = false;
        document.getElementById('cdn_record_domain').disabled = false;
        document.getElementById('cdn_record_type').disabled = false;

        openModal('cdnRecordModal', 'cdnRecordModalTitle', '添加 CDN记录');
    }

}

// 选择配置 config_id
function onCdnRecordConfigSelect(sel) {
    const opt = sel.options[sel.selectedIndex];
    // console.log("onCdnRecordConfigSelect", opt);
}

// 选择类型 type
function onCdnRecordTypeSelect(sel) {
    const opt = sel.options[sel.selectedIndex];
    // console.log("onCdnRecordTypeSelect", opt);
}

// 选择回源协议 origin_protocol
function onCdnRecordOriginProtocolSelect(sel) {
    const opt = sel.options[sel.selectedIndex];
    // console.log("onCdnRecordOriginProtocolSelect", opt);
}

// 选择IP组 ip_group_id
function onCdnRecordIpGroupSelect(sel) {
    const opt = sel.options[sel.selectedIndex];
    // console.log("onCdnRecordIpGroupSelect", opt);
}

// CDN记录管理 -> 表单提交
document.getElementById('cdnRecordForm').onsubmit = async (e) => {
    e.preventDefault();

    // 表单数据
    const form_id = Number(document.getElementById('cdn_record_id').value.trim());
    const form_data = {
        config_id: Number(document.getElementById('cdn_record_config_id').value.trim()),
        domain: document.getElementById('cdn_record_domain').value.trim(),
        type: document.getElementById('cdn_record_type').value.trim(),
        origin_protocol: document.getElementById('cdn_record_origin_protocol').value.trim(),
        http_port: Number(document.getElementById('cdn_record_http_port').value.trim()),
        https_port: Number(document.getElementById('cdn_record_https_port').value.trim()),
        ip_group_id: Number(document.getElementById('cdn_record_ip_group_id').value.trim()),
    };
    console.log("form_id, form_data", form_id, form_data);

    // 校验
    if (isNaN(form_data.config_id) || form_data.config_id <= 0) {
        return alert('请选择配置')
    };
    if (!form_data.domain) {
        return alert('请输入域名')
    };
    if (!form_data.type) {
        return alert('请选择记录类型')
    };
    if (isNaN(form_data.ip_group_id) || form_data.ip_group_id <= 0) {
        return alert('请选择IP组')
    };

    // 提交
    if (isNaN(form_id) || form_id <= 0) {
        await API.post('/api/cdn-records', form_data);
    } else {
        await API.patch(`/api/cdn-records/${form_id}`, form_data);
    }

    // 刷新页面
    location.reload();
};

// 删除
async function deleteCdnRecord(id) {
    if (confirm('确定删除此记录？')) {
        await API.delete(`/api/cdn-records/${id}`);

        // 刷新页面
        location.reload();
    }
}

// --- CDN记录管理 --- 
// --- DNS记录管理 --- 

// 状态
async function toggleDnsRecordStatus(id, checkbox) {
    await API.patch(`/api/dns-records/${id}/active`, { is_active: checkbox.checked });
}

// DNS记录管理（模态框）
async function openDnsRecordModal(id) {

    // IP 组列表
    const { data: ipGroups } = await API.get(`/api/ip-groups`);
    console.log("ipGroups", ipGroups);
    if (ipGroups) {
        const ipGroupSelect = document.getElementById('dns_record_ip_group_id');
        ipGroupSelect.innerHTML = '';
        addPlaceholder(ipGroupSelect, '请选择IP组');
        ipGroups.forEach(d => {
            const opt = document.createElement('option');
            opt.value = d.id;
            opt.text = `${d.name}`;
            ipGroupSelect.appendChild(opt);
        });
    }

    // CDN列表
    const { data: cdnRecords } = await API.get(`/api/cdn-records`);
    console.log("cdnRecords", cdnRecords);
    if (cdnRecords) {
        const cdnRecordSelect = document.getElementById('dns_record_cdn_record_id');
        cdnRecordSelect.innerHTML = '';
        addPlaceholder(cdnRecordSelect, '请选择CDN记录');
        cdnRecords.forEach(d => {
            const opt = document.createElement('option');
            opt.value = d.id;
            opt.text = `${d.domain}`;
            cdnRecordSelect.appendChild(opt);
        });
    }

    // 编辑
    if (id) {
        const { data: dnsRecord } = await API.get(`/api/dns-records/${id}`);
        console.log("dnsRecord", dnsRecord);
        if (!dnsRecord) {
            alert('DNS记录数据异常');
            return;
        }

        document.getElementById('dns_record_id').value = dnsRecord.id;
        document.getElementById('dns_record_id').disabled = true;
        document.getElementById('dns_record_config_id').value = dnsRecord.config_id;
        document.getElementById('dns_record_config_id').disabled = true;
        document.getElementById('dns_record_domain').value = dnsRecord.domain;
        document.getElementById('dns_record_domain').disabled = true;
        document.getElementById('dns_record_type').value = dnsRecord.type;
        document.getElementById('dns_record_type').disabled = true;

        if (dnsRecord.type === 'CNAME') {
            document.getElementById('dns_record_cdn_record_id').value = dnsRecord.cdn_record_id;
            document.getElementById('dns_record_cdn_record_id').disabled = false;
        } else {
            document.getElementById('dns_record_ip_group_id').value = dnsRecord.ip_group_id;
            document.getElementById('dns_record_ip_group_id').disabled = false;
        }

        openModal('dnsRecordModal', 'dnsRecordModalTitle', '编辑 DNS记录');
    }

    // 创建
    if (!id) {
        document.getElementById('dns_record_id').value = 0;
        document.getElementById('dns_record_id').disabled = true;
        document.getElementById('dns_record_config_id').disabled = false;
        document.getElementById('dns_record_domain').disabled = false;
        document.getElementById('dns_record_type').disabled = false;
        document.getElementById('dns_record_cdn_record_id').disabled = true;
        document.getElementById('dns_record_ip_group_id').disabled = true;

        openModal('dnsRecordModal', 'dnsRecordModalTitle', '添加 DNS记录');
    }

}

// 选择配置 config_id
function onDnsRecordConfigSelect(sel) {
    const opt = sel.options[sel.selectedIndex];
    // console.log("onCdnRecordIponDnsRecordConfigSelectGroupSelect", opt);
}

// 选择类型 type
function onDnsRecordTypeSelect(sel) {
    const opt = sel.options[sel.selectedIndex];
    // console.log("onDnsRecordTypeSelect", opt);
    if (opt.value === 'CNAME') {
        document.getElementById('dns_record_ip_group_id').disabled = true;
        document.getElementById('dns_record_cdn_record_id').disabled = false;
        document.getElementById('dns_record_ip_group_id').value = '';
    } else {
        document.getElementById('dns_record_ip_group_id').disabled = false;
        document.getElementById('dns_record_cdn_record_id').disabled = true;
        document.getElementById('dns_record_cdn_record_id').value = '';
    }
}

// 选择IP组 ip_group_id
function onDnsRecordIpGroupSelect(sel) {
    const opt = sel.options[sel.selectedIndex];
    // console.log("onDnsRecordIpGroupSelect", opt);
}

// 选择CDN记录 cdn_record_id
function onDnsRecordCdnRecordSelect(sel) {
    const opt = sel.options[sel.selectedIndex];
    // console.log("onDnsRecordCdnRecordSelect", opt);
}

// DNS记录管理（模态框） -> 表单提交
document.getElementById('dnsRecordForm').onsubmit = async (e) => {
    e.preventDefault();

    // 表单数据
    const form_id = Number(document.getElementById('dns_record_id').value.trim());
    const form_data = {
        config_id: Number(document.getElementById('dns_record_config_id').value.trim()),
        domain: document.getElementById('dns_record_domain').value.trim(),
        type: document.getElementById('dns_record_type').value.trim(),
        ip_group_id: Number(document.getElementById('dns_record_ip_group_id').value.trim()),
        cdn_record_id: Number(document.getElementById('dns_record_cdn_record_id').value.trim()),
    };
    console.log("form_id, form_data", form_id, form_data);

    // 校验
    if (isNaN(form_data.config_id) || form_data.config_id <= 0) {
        return alert('请选择配置')
    };
    if (!form_data.domain) {
        return alert('请输入域名')
    };
    if (!form_data.type) {
        return alert('请选择记录类型')
    };
    if (form_data.type === 'CNAME') {
        if (isNaN(form_data.cdn_record_id) || form_data.cdn_record_id <= 0) {
            return alert('请选择CDN记录')
        };
    } else {
        if (isNaN(form_data.ip_group_id) || form_data.ip_group_id <= 0) {
            return alert('请选择IP组')
        };
    }

    // 提交
    if (isNaN(form_id) || form_id <= 0) {
        await API.post('/api/dns-records', form_data);
    } else {
        await API.patch(`/api/dns-records/${form_id}`, form_data);
    }

    // 刷新页面
    location.reload();
};

// 删除
async function deleteDnsRecord(id) {
    if (confirm('确定删除此记录？')) {
        await API.delete(`/api/dns-records/${id}`);

        // 刷新页面
        location.reload();
    }
}

// --- DNS记录管理 --- 
// --- 设备实时上报 --- 

// 管理设备（模态框）
async function openDeviceModal(id) {
    document.getElementById('deviceForm').reset();

    // 编辑
    if (id) {
        const { data: device } = await API.get(`/api/devices/${id}`);
        console.log("device", device);
        if (!device) {
            alert('设备数据异常');
            return;
        }

        document.getElementById('device_id').value = device.id;
        document.getElementById('device_id').disabled = true;
        document.getElementById('device_code').value = device.code;
        document.getElementById('device_code').disabled = true;
        document.getElementById('device_name').value = device.name

        openModal('deviceModal', 'deviceModalTitle', '编辑 设备');
    }

    // 创建
    if (!id) {
        document.getElementById('device_id').value = '';
        document.getElementById('device_id').disabled = true;

        const deviceCode = uuid.v7();
        if (!deviceCode) {
            alert('生成设备编码失败');
            return;
        }
        document.getElementById('device_code').value = deviceCode;
        document.getElementById('device_code').disabled = false;

        openModal('deviceModal', 'deviceModalTitle', '添加 设备');
    }

}

// 管理设备（模态框） -> 表单提交
document.getElementById('deviceForm').onsubmit = async (e) => {
    e.preventDefault();

    // 表单数据
    const form_id = Number(document.getElementById('device_id').value.trim());
    const form_data = {
        code: document.getElementById('device_code').value.trim(),
        name: document.getElementById('device_name').value.trim(),
    };
    console.log("form_id, form_data", form_id, form_data);

    // 校验
    if (!form_data.code) {
        return alert('请输入设备编码')
    };
    if (!form_data.name) {
        return alert('请输入设备名称')
    };

    // 提交
    if (isNaN(form_id) || form_id <= 0) {
        await API.post('/api/devices', form_data);
    } else {
        await API.patch(`/api/devices/${form_id}`, form_data);
    }

    // 刷新页面
    location.reload();
};

// 删除
async function deleteDevice(id) {
    if (confirm('确定删除此设备？')) {
        await API.delete(`/api/devices/${id}`);

        // 刷新页面
        location.reload();
    }
}

// 实时 设备 IP 历史 弹窗
async function viewLiveDeviceIpHistory(id, name) {
    const { data: deviceIpHistory } = await API.get(`/api/devices/${id}/ip-history`);
    console.log("deviceIpHistory", deviceIpHistory);
    if (!deviceIpHistory) {
        alert('设备数据异常');
        return;
    }
    if (deviceIpHistory.length === 0) {
        alert('设备无历史IP数据');
        return;
    }

    // 填充代码块
    const pre = document.getElementById('liveDeviceIpHistoryPre');

    // 格式化 IP 列表
    const ipList = deviceIpHistory.map(d => `${fmtDate(d.created_at)} - ${d.ip_v4 || '-'} / ${d.ip_v6 || '-'}`);

    pre.textContent = ipList.join('\n');

    // 更新计数
    document.getElementById('liveDeviceIpHistoryCount').textContent = deviceIpHistory.length;

    // 打开弹窗（统一走你已有的 openModal 即可）
    openModal('liveDeviceIpHistoryModal', 'liveDeviceIpHistoryTitle', `实时 设备 IP 历史 - ${name}`);
}

document.addEventListener('DOMContentLoaded', function () {
    const now = Math.floor(Date.now() / 1000);
    // console.log("当前时间戳（秒）:", now);
    const tenMinutes = 12 * 60;

    document.querySelectorAll('#deviceTable tr[data-last-heartbeat]').forEach(row => {
        const heartbeatAt = parseInt(row.getAttribute('data-last-heartbeat'));
        // console.log("设备心跳时间戳:", heartbeatAt, "差值:", now - heartbeatAt);

        // 如果是零值（从未心跳），heartbeatAt <= 0
        if (heartbeatAt <= 0) {
            row.classList.add('never-heartbeat');
            return;
        }

        if (now - heartbeatAt > tenMinutes) {
            row.classList.add('offline');
        } else {
            row.classList.remove('offline');
        }
    });
});

// --- 设备实时上报 --- 
// --- IP 组 --- 

// 选择设备 devices
function onIpGroupDevicesSelect(sel) {
    const opt = sel.options[sel.selectedIndex];
    // console.log("onIpGroupDevicesSelect", opt);
}

// 管理组（模态框）
async function openIpGroupModal(id) {
    document.getElementById('ipGroupForm').reset();

    // 设备列表
    const { data: devices } = await API.get(`/api/devices`);
    console.log("devices", devices);
    if (devices) {
        const deviceSelect = document.getElementById('ip_group_devices');
        deviceSelect.innerHTML = '';
        devices.forEach(d => {
            const opt = document.createElement('option');
            opt.value = d.id;
            opt.text = `${d.name} - ${d.ip_v4 || '-'} / ${d.ip_v6 || '-'}`;
            deviceSelect.appendChild(opt);
        });
    }

    // 编辑
    if (id) {
        const { data: ipGroup } = await API.get(`/api/ip-groups/${id}`);
        console.log("ipGroup", ipGroup);
        if (!ipGroup) {
            alert('IP 组数据异常');
            return;
        }

        document.getElementById('ip_group_id').value = ipGroup.id;
        document.getElementById('ip_group_id').disabled = true;
        document.getElementById('ip_group_name').value = ipGroup.name;
        document.getElementById('ip_group_description').value = ipGroup.description || '';

        // 高亮已选
        if (ipGroup.devices && ipGroup.devices.length > 0) {
            const deviceSelect = document.getElementById('ip_group_devices');
            Array.from(deviceSelect.options).forEach(opt => opt.selected = ipGroup.devices.includes(Number(opt.value)));
        }

        if (ipGroup.description == window.deviceCode) {
            document.getElementById('ip_group_description').disabled = true;
        } else {
            document.getElementById('ip_group_description').disabled = false;
        }

        openModal('ipGroupModal', 'ipGroupModalTitle', '编辑 IP 组');
    }

    // 创建
    if (!id) {
        document.getElementById('ip_group_id').value = '';
        document.getElementById('ip_group_id').disabled = true;
        document.getElementById('ip_group_description').disabled = false;

        openModal('ipGroupModal', 'ipGroupModalTitle', '添加 IP 组');
    }

}

// 管理组（模态框） -> 表单提交
document.getElementById('ipGroupForm').onsubmit = async (e) => {
    e.preventDefault();

    // 表单数据
    const form_id = Number(document.getElementById('ip_group_id').value.trim());
    const form_data = {
        name: document.getElementById('ip_group_name').value.trim(),
        description: document.getElementById('ip_group_description').value.trim(),
        devices: Array.from(document.getElementById('ip_group_devices').selectedOptions).map(opt => Number(opt.value.trim())),
    };
    console.log("form_id, form_data", form_id, form_data);

    // 校验
    if (!form_data.name) {
        return alert('请输入组名')
    };
    if (form_data.devices.length === 0) {
        return alert('请选择设备')
    };

    // 提交
    if (isNaN(form_id) || form_id <= 0) {
        await API.post('/api/ip-groups', form_data);
    } else {
        await API.patch(`/api/ip-groups/${form_id}`, form_data);
    }

    // 刷新页面
    location.reload();
};

// 删除
async function deleteIpGroup(id) {
    if (confirm('确定删除此 IP 组？')) {
        await API.delete(`/api/ip-groups/${id}`);

        // 刷新页面
        location.reload();
    }
}

// 实时 设备 IP 地址 弹窗
async function viewLiveDeviceIpAddresse(id, name) {
    const { data: ipGroupDevices } = await API.get(`/api/ip-groups/${id}/devices`);
    console.log("ipGroupDevices", ipGroupDevices);
    if (!ipGroupDevices) {
        alert('IP组设备数据异常');
        return;
    }
    if (ipGroupDevices.length === 0) {
        alert('当前组内设备无数据');
        return;
    }

    // 填充代码块
    const pre = document.getElementById('liveDeviceIpAddressePre');

    // 格式化 IP 列表
    const ipList = ipGroupDevices.map(d => `${d.name} - ${d.ip_v4 || '-'} / ${d.ip_v6 || '-'}`);

    pre.textContent = ipList.join('\n');

    // 更新计数
    document.getElementById('liveDeviceIpAddresseCount').textContent = ipGroupDevices.length;

    // 打开弹窗（统一走你已有的 openModal 即可）
    openModal('liveDeviceIpAddresseModal', 'liveDeviceIpAddresseTitle', `实时 设备 IP 地址 - ${name}`);
}

// --- IP 组 ---
// --- 通知（WebHook） --- 

// 选择平台类型 platform
function onWebhookPlatformSelect(sel) {
    const opt = sel.options[sel.selectedIndex];
    // console.log("onWebhookPlatformSelect", opt);
    const webhook_id = document.getElementById('webhook_id').value.trim()
    // console.log("onWebhookPlatformSelect webhook_id", webhook_id);
    if (opt.value === 'custom_http') {
        document.getElementById('webhook_url').disabled = false;
        document.getElementById('webhook_url').value = '';
        document.getElementById('webhook_url').required = true;
        document.getElementById('webhook_url_group').style.display = 'block';
    } else {
        document.getElementById('webhook_url').disabled = true;
        document.getElementById('webhook_url').value = '';
        document.getElementById('webhook_url').required = false;
        document.getElementById('webhook_url_group').style.display = 'none';
    }
    if (!webhook_id) {
        switch (opt.value) {
            case 'custom_http':
                document.getElementById('webhook_template_title').value = '[IPVibe] [[.DeviceName]] IP 更新'
                document.getElementById('webhook_template_content').value = `事件类型: {{.EventType}}
事件时间: {{.EventTime}}
事件内容: {{.EventBody}}
设备编号: {{.DeviceCode}}
设备名称: {{.DeviceName}}
IPV4地址: {{.IPv4Addr}}
IPV6地址: {{.IPv6Addr}}
记录域名: {{.RecordDomain}}
记录类型: {{.RecordType}}
平台名称: {{.PlatformName}}
平台类型: {{.PlatformType}}`
                break;
        }
    }
}

// 修改通知状态
async function toggleWebhookStatus(id, checkbox) {
    await API.patch(`/api/webhooks/${id}/active`, { is_active: checkbox.checked });
}

// 管理通知（模态框）
async function openWebhookModal(id) {
    document.getElementById('webhookForm').reset();

    // 编辑
    if (id) {
        const { data: webhook } = await API.get(`/api/webhooks/${id}`);
        console.log("webhook", webhook);
        if (!webhook) {
            alert('通知数据异常');
            return;
        }

        document.getElementById('webhook_id').value = webhook.id;
        document.getElementById('webhook_id').disabled = true;
        document.getElementById('webhook_name').value = webhook.name;
        document.getElementById('webhook_platform').value = webhook.platform;
        document.getElementById('webhook_platform').disabled = true;
        document.getElementById('webhook_url').value = webhook.url;
        document.getElementById('webhook_secret').value = webhook.secret;
        document.getElementById('webhook_template_title').value = webhook.template_title;
        document.getElementById('webhook_template_content').value = webhook.template_content;

        if (webhook.platform === 'custom_http') {
            document.getElementById('webhook_url').disabled = false;
            document.getElementById('webhook_url').required = true;
            document.getElementById('webhook_url_group').style.display = 'block';
        } else {
            document.getElementById('webhook_url').disabled = true;
            document.getElementById('webhook_url').required = false;
            document.getElementById('webhook_url_group').style.display = 'none';
        }

        openModal('webhookModal', 'webhookModalTitle', '编辑 通知');
    }

    // 创建
    if (!id) {
        document.getElementById('webhook_id').value = '';
        document.getElementById('webhook_id').disabled = true;

        openModal('webhookModal', 'webhookModalTitle', '添加 通知');
    }

}

// 管理通知（模态框） -> 表单提交
document.getElementById('webhookForm').onsubmit = async (e) => {
    e.preventDefault();

    // 表单数据
    const form_id = Number(document.getElementById('webhook_id').value.trim());
    const form_data = {
        name: document.getElementById('webhook_name').value.trim(),
        platform: document.getElementById('webhook_platform').value.trim(),
        url: document.getElementById('webhook_url').value.trim(),
        secret: document.getElementById('webhook_secret').value.trim(),
        template_title: document.getElementById('webhook_template_title').value.trim(),
        template_content: document.getElementById('webhook_template_content').value.trim(),
    };
    console.log("form_id, form_data", form_id, form_data);

    // 校验
    if (!form_data.name) {
        return alert('请输入组名')
    };
    if (!form_data.platform) {
        return alert('请选择平台类型')
    };
    if (!form_data.template_title) {
        return alert('请输入通知标题模板')
    };
    if (!form_data.template_content) {
        return alert('请输入通知内容模板')
    };

    // 提交
    if (isNaN(form_id) || form_id <= 0) {
        await API.post('/api/webhooks', form_data);
    } else {
        await API.patch(`/api/webhooks/${form_id}`, form_data);
    }

    // 刷新页面
    location.reload();
};

// 删除通知
async function deleteWebhook(id) {
    if (confirm('确定删除此通知？')) {
        await API.delete(`/api/webhooks/${id}`);

        // 刷新页面
        location.reload();
    }
}

// --- 通知（WebHook） --- 
// --- 通知订阅 --- 

// 选择通知 webhook_id
function onWebhookSubscriptionWebhookIdSelect(sel) {
    const opt = sel.options[sel.selectedIndex];
    // console.log("onWebhookSubscriptionWebhookIdSelect", opt);
}

// 选择订阅对象 obj_id
function onWebhookSubscriptionObjIdSelect(sel) {
    const opt = sel.options[sel.selectedIndex];
    // console.log("onWebhookSubscriptionObjIdSelect", opt);
}

// 修改通知订阅状态
async function toggleWebhookSubscriptionStatus(id, checkbox) {
    await API.patch(`/api/webhook-subscriptions/${id}/active`, { is_active: checkbox.checked });
}

// 管理通知订阅（模态框）
async function openWebhookSubscriptionModal(id, obj_type) {
    document.getElementById('webhookSubscriptionForm').reset();

    // 通知列表
    const { data: webhooks } = await API.get(`/api/webhooks`);
    console.log("webhooks", webhooks);
    if (webhooks) {
        const webhookIdSelect = document.getElementById('webhook_subscription_webhook_id');
        webhookIdSelect.innerHTML = '';
        addPlaceholder(webhookIdSelect, '请选择通知');
        webhooks.forEach(d => {
            const opt = document.createElement('option');
            opt.value = d.id;
            opt.text = `${d.name}`;
            webhookIdSelect.appendChild(opt);
        });
    }

    var modalTitle = ''
    switch (obj_type) {
        case 'device':
            modalTitle = '设备订阅';

            // 设备列表
            const { data: devices } = await API.get(`/api/devices`);
            console.log("devices", devices);
            if (devices) {
                const objIdSelect = document.getElementById('webhook_subscription_obj_id');
                objIdSelect.innerHTML = '';
                addPlaceholder(objIdSelect, '请选择设备');
                devices.forEach(d => {
                    const opt = document.createElement('option');
                    opt.value = d.id;
                    opt.text = `${d.name}`;
                    objIdSelect.appendChild(opt);
                });
            }

            break;
        case 'dns_record':
            modalTitle = 'DNS订阅';

            // DNS列表
            const { data: dnsRecords } = await API.get(`/api/dns-records`);
            console.log("dnsRecords", dnsRecords);
            if (dnsRecords) {
                const objIdSelect = document.getElementById('webhook_subscription_obj_id');
                objIdSelect.innerHTML = '';
                addPlaceholder(objIdSelect, '请选择DNS记录');
                dnsRecords.forEach(d => {
                    const opt = document.createElement('option');
                    opt.value = d.id;
                    opt.text = `${d.domain} - ${d.type}`;
                    objIdSelect.appendChild(opt);
                });
            }

            break;
        case 'cdn_record':
            modalTitle = 'CDN订阅';

            // CDN列表
            const { data: cdnRecords } = await API.get(`/api/cdn-records`);
            console.log("cdnRecords", cdnRecords);
            if (cdnRecords) {
                const objIdSelect = document.getElementById('webhook_subscription_obj_id');
                objIdSelect.innerHTML = '';
                addPlaceholder(objIdSelect, '请选择CDN记录');
                cdnRecords.forEach(d => {
                    const opt = document.createElement('option');
                    opt.value = d.id;
                    opt.text = `${d.domain} - ${d.type}`;
                    objIdSelect.appendChild(opt);
                });
            }

            break;
        default:
            return alert('请选择订阅对象类型')
    }

    // 编辑
    if (id) {
        const { data: webhookSubscription } = await API.get(`/api/webhook-subscriptions/${id}`);
        console.log("webhookSubscription", webhookSubscription);
        if (!webhookSubscription) {
            alert('通知订阅数据异常');
            return;
        }

        document.getElementById('webhook_subscription_id').value = webhookSubscription.id;
        document.getElementById('webhook_subscription_id').disabled = true;
        document.getElementById('webhook_subscription_name').value = webhookSubscription.name;
        document.getElementById('webhook_subscription_webhook_id').value = webhookSubscription.webhook_id;
        document.getElementById('webhook_subscription_webhook_id').disabled = true;
        document.getElementById('webhook_subscription_obj_id').value = webhookSubscription.obj_id;
        document.getElementById('webhook_subscription_obj_id').disabled = true;
        document.getElementById('webhook_subscription_obj_type').value = webhookSubscription.obj_type;
        document.getElementById('webhook_subscription_obj_type').disabled = true;

        openModal('webhookSubscriptionModal', 'webhookSubscriptionModalTitle', `编辑 ${modalTitle}`);
    }

    // 创建
    if (!id) {
        document.getElementById('webhook_subscription_id').value = '';
        document.getElementById('webhook_subscription_id').disabled = true;
        document.getElementById('webhook_subscription_webhook_id').disabled = false;
        document.getElementById('webhook_subscription_obj_id').disabled = false;
        document.getElementById('webhook_subscription_obj_type').value = obj_type;
        document.getElementById('webhook_subscription_obj_type').disabled = true;

        openModal('webhookSubscriptionModal', 'webhookSubscriptionModalTitle', `添加 ${modalTitle}`);
    }

}

// 管理通知订阅（模态框） -> 表单提交
document.getElementById('webhookSubscriptionForm').onsubmit = async (e) => {
    e.preventDefault();

    // 表单数据
    const form_id = Number(document.getElementById('webhook_subscription_id').value.trim());
    const form_data = {
        name: document.getElementById('webhook_subscription_name').value.trim(),
        webhook_id: Number(document.getElementById('webhook_subscription_webhook_id').value.trim()),
        obj_type: document.getElementById('webhook_subscription_obj_type').value.trim(),
        obj_id: Number(document.getElementById('webhook_subscription_obj_id').value.trim()),
    };
    console.log("form_id, form_data", form_id, form_data);

    // 校验
    if (!form_data.name) {
        return alert('请输入名称')
    };
    if (isNaN(form_data.webhook_id) || form_data.webhook_id <= 0) {
        return alert('请选择通知渠道')
    };
    if (!form_data.obj_type) {
        return alert('请选择订阅对象类型')
    };
    if (isNaN(form_data.obj_id) || form_data.obj_id <= 0) {
        return alert('请选择订阅对象')
    };

    // 提交
    if (isNaN(form_id) || form_id <= 0) {
        await API.post('/api/webhook-subscriptions', form_data);
    } else {
        await API.patch(`/api/webhook-subscriptions/${form_id}`, form_data);
    }

    // 刷新页面
    location.reload();
};

// 删除通知订阅
async function deleteWebhookSubscription(id) {
    if (confirm('确定删除此通知订阅？')) {
        await API.delete(`/api/webhook-subscriptions/${id}`);

        // 刷新页面
        location.reload();
    }
}

// --- 通知订阅 --- 
// --- 任务管理 ---

// 启动全部任务
async function startTask() {
    if (confirm('确定启动全部任务？')) {
        await API.post(`/api/tasks/start`);

        // 刷新页面
        location.reload();
    }
}

// 停止全部任务
async function stopTask() {
    if (confirm('确定停止全部任务？')) {
        await API.post(`/api/tasks/stop`);

        // 刷新页面
        location.reload();
    }
}

// 清理全部任务
async function cleanTask() {
    if (confirm('确定清理全部任务？')) {
        await API.post(`/api/tasks/clean`);

        // 刷新页面
        location.reload();
    }
}

// 运行某任务
async function runTask(id) {
    if (confirm('确定运行此任务？')) {
        await API.post(`/api/tasks/${id}`);

        // 刷新页面
        location.reload();
    }
}

// 实时 执行时间 弹窗
async function viewLiveExecutionTime(id, name) {
    const { data: task } = await API.get(`/api/tasks/${id}`);
    console.log("task", task);
    if (!task) {
        alert('任务数据异常');
        return;
    }
    if (task.length === 0) {
        alert('当前任务执行时间无数据');
        return;
    }

    // 填充代码块
    const pre = document.getElementById('liveExecutionTimePre');

    pre.textContent = task.next_runs.join('\n');

    // 更新计数
    document.getElementById('liveExecutionTimeCount').textContent = task.next_runs.length;

    // 打开弹窗（统一走你已有的 openModal 即可）
    openModal('liveExecutionTimeModal', 'liveExecutionTimeTitle', `实时 执行时间 - ${name}`);
}

// --- 任务管理 ---