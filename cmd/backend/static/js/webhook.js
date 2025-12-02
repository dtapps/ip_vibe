
function changeWebhookPlatform(platform) {
    switch (platform) {
        case 'wechat_work':
            document.getElementById('webhook_url').disabled = false;
            document.getElementById('webhook_url').required = true;
            document.getElementById('webhook_url_group').style.display = 'block';

            document.getElementById('webhook_secret').disabled = false;
            document.getElementById('webhook_secret').required = false;
            document.getElementById('webhook_secret_group').style.display = 'none';
            break;
        case 'dingtalk':
            document.getElementById('webhook_url').disabled = false;
            document.getElementById('webhook_url').required = true;
            document.getElementById('webhook_url_group').style.display = 'block';

            document.getElementById('webhook_secret').disabled = false;
            document.getElementById('webhook_secret').required = false;
            document.getElementById('webhook_secret_group').style.display = 'block';
            break;
        case 'feishu':
            document.getElementById('webhook_url').disabled = false;
            document.getElementById('webhook_url').required = true;
            document.getElementById('webhook_url_group').style.display = 'block';

            document.getElementById('webhook_secret').disabled = false;
            document.getElementById('webhook_secret').required = false;
            document.getElementById('webhook_secret_group').style.display = 'block';
            break;
        case 'bark':
            document.getElementById('webhook_url').disabled = false;
            document.getElementById('webhook_url').required = false;
            document.getElementById('webhook_url_group').style.display = 'block';

            document.getElementById('webhook_secret').disabled = false;
            document.getElementById('webhook_secret').required = false;
            document.getElementById('webhook_secret_group').style.display = 'block';
            break;
        case 'ft07':
            document.getElementById('webhook_url').disabled = false;
            document.getElementById('webhook_url').required = false;
            document.getElementById('webhook_url_group').style.display = 'block';

            document.getElementById('webhook_secret').disabled = false;
            document.getElementById('webhook_secret').required = true;
            document.getElementById('webhook_secret_group').style.display = 'block';
            break;
        case 'push_deer':
            document.getElementById('webhook_url').disabled = false;
            document.getElementById('webhook_url').required = false;
            document.getElementById('webhook_url_group').style.display = 'block';

            document.getElementById('webhook_secret').disabled = false;
            document.getElementById('webhook_secret').required = true;
            document.getElementById('webhook_secret_group').style.display = 'block';
            break;
        case 'custom_http':
            document.getElementById('webhook_url').disabled = false;
            document.getElementById('webhook_url').required = true;
            document.getElementById('webhook_url_group').style.display = 'block';

            document.getElementById('webhook_secret').disabled = true;
            document.getElementById('webhook_secret').required = false;
            document.getElementById('webhook_secret_group').style.display = 'none';
            break;
    }
}

// 选择平台类型 platform
function onWebhookPlatformSelect(sel) {
    const opt = sel.options[sel.selectedIndex];
    // console.log("onWebhookPlatformSelect", opt);
    const webhook_id = document.getElementById('webhook_id').value.trim()
    // console.log("onWebhookPlatformSelect webhook_id", webhook_id);

    changeWebhookPlatform(opt.value);

    if (!webhook_id) {
        document.getElementById('webhook_template_title').value = '[IPVibe] [[.DeviceName]] IP 更新'
        switch (opt.value) {
            case 'wechat_work':
                document.getElementById('webhook_template_content').value = `事件类型: [[.EventType]]
事件时间: [[.EventTime]]
事件内容: [[.EventBody]]
设备编号: [[.DeviceCode]]
设备名称: [[.DeviceName]]
IPV4地址: [[.IPv4Addr]]
IPV6地址: [[.IPv6Addr]]
记录域名: [[.RecordDomain]]
记录类型: [[.RecordType]]
平台名称: [[.PlatformName]]
平台类型: [[.PlatformType]]`
                break;
            case 'dingtalk':
                document.getElementById('webhook_template_content').value = `事件类型: [[.EventType]]
事件时间: [[.EventTime]]
事件内容: [[.EventBody]]
设备编号: [[.DeviceCode]]
设备名称: [[.DeviceName]]
IPV4地址: [[.IPv4Addr]]
IPV6地址: [[.IPv6Addr]]
记录域名: [[.RecordDomain]]
记录类型: [[.RecordType]]
平台名称: [[.PlatformName]]
平台类型: [[.PlatformType]]`
                break;
            case 'feishu':
                document.getElementById('webhook_template_content').value = `事件类型: [[.EventType]]
事件时间: [[.EventTime]]
事件内容: [[.EventBody]]
设备编号: [[.DeviceCode]]
设备名称: [[.DeviceName]]
IPV4地址: [[.IPv4Addr]]
IPV6地址: [[.IPv6Addr]]
记录域名: [[.RecordDomain]]
记录类型: [[.RecordType]]
平台名称: [[.PlatformName]]
平台类型: [[.PlatformType]]`
                break;
            case 'bark':
                document.getElementById('webhook_template_content').value = `事件类型: [[.EventType]]
事件时间: [[.EventTime]]
事件内容: [[.EventBody]]
设备编号: [[.DeviceCode]]
设备名称: [[.DeviceName]]
IPV4地址: [[.IPv4Addr]]
IPV6地址: [[.IPv6Addr]]
记录域名: [[.RecordDomain]]
记录类型: [[.RecordType]]
平台名称: [[.PlatformName]]
平台类型: [[.PlatformType]]`
                break;
            case 'ft07':
                document.getElementById('webhook_template_content').value = `事件类型: [[.EventType]]
事件时间: [[.EventTime]]
事件内容: [[.EventBody]]
设备编号: [[.DeviceCode]]
设备名称: [[.DeviceName]]
IPV4地址: [[.IPv4Addr]]
IPV6地址: [[.IPv6Addr]]
记录域名: [[.RecordDomain]]
记录类型: [[.RecordType]]
平台名称: [[.PlatformName]]
平台类型: [[.PlatformType]]`
                break;
            case 'push_deer':
                document.getElementById('webhook_template_content').value = `事件类型: [[.EventType]]
事件时间: [[.EventTime]]
事件内容: [[.EventBody]]
设备编号: [[.DeviceCode]]
设备名称: [[.DeviceName]]
IPV4地址: [[.IPv4Addr]]
IPV6地址: [[.IPv6Addr]]
记录域名: [[.RecordDomain]]
记录类型: [[.RecordType]]
平台名称: [[.PlatformName]]
平台类型: [[.PlatformType]]`
                break;
            case 'custom_http':
                document.getElementById('webhook_template_content').value = `事件类型: [[.EventType]]
事件时间: [[.EventTime]]
事件内容: [[.EventBody]]
设备编号: [[.DeviceCode]]
设备名称: [[.DeviceName]]
IPV4地址: [[.IPv4Addr]]
IPV6地址: [[.IPv6Addr]]
记录域名: [[.RecordDomain]]
记录类型: [[.RecordType]]
平台名称: [[.PlatformName]]
平台类型: [[.PlatformType]]`
                break;
        }
    }
}

// 修改通知状态
async function toggleWebhookStatus(id, checkbox) {
    try {
        await API.patch(`/api/webhooks/${id}/active`, { is_active: checkbox.checked });
    } catch (error) {
        // 恢复checkbox状态
        checkbox.checked = !checkbox.checked;
        alert(`更新状态失败: ${error.message}`);
        console.error("更新通知状态失败", error);
    }
}

// 管理通知（模态框）
async function openWebhookModal(id) {
    document.getElementById('webhookForm').reset();

    try {
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

            changeWebhookPlatform(webhook.platform);

            openModal('webhookModal', 'webhookModalTitle', '编辑 通知');
        }

        // 创建
        if (!id) {
            document.getElementById('webhook_id').value = '';
            document.getElementById('webhook_id').disabled = true;

            openModal('webhookModal', 'webhookModalTitle', '添加 通知');
        }
    } catch (error) {
        alert(`加载通知信息失败: ${error.message}`);
        console.error("打开通知模态框失败", error);
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

    try {
        // 提交
        if (isNaN(form_id) || form_id <= 0) {
            await API.post('/api/webhooks', form_data);
        } else {
            await API.patch(`/api/webhooks/${form_id}`, form_data);
        }

        // 刷新页面
        location.reload();
    } catch (error) {
        alert(`操作失败: ${error.message}`);
        console.error("通知操作失败", error);
    }
};

// 删除通知
async function deleteWebhook(id) {
    if (confirm('确定删除此通知？')) {
        try {
            await API.delete(`/api/webhooks/${id}`);
            // 刷新页面
            location.reload();
        } catch (error) {
            alert(`删除失败: ${error.message}`);
            console.error("删除通知失败", error);
        }
    }
}
