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

    try {
        // 提交
        if (isNaN(form_id) || form_id <= 0) {
            await API.post('/api/configs', form_data);
        } else {
            await API.patch(`/api/configs/${form_id}`, form_data);
        }

        // 刷新页面
        location.reload();
    } catch (error) {
        alert(`操作失败: ${error.message}`);
        console.error("配置操作失败", error);
    }
};

// 删除
async function deleteConfig(id) {
    if (confirm('确定删除此配置？相关记录也会被删除')) {
        try {
            await API.delete(`/api/configs/${id}`);
            // 刷新页面
            location.reload();
        } catch (error) {
            alert(`删除失败: ${error.message}`);
            console.error("删除配置失败", error);
        }
    }
}
