// 选择协议 protocols
function onPortForwardProtocolsSelect(sel) {
    const opt = sel.options[sel.selectedIndex];
    // console.log("onPortForwardProtocolsSelect", opt);
}

// 选择设备 device_id
function onPortForwardDeviceIdSelect(sel) {
    const opt = sel.options[sel.selectedIndex];
    // console.log("onPortForwardDeviceIdSelect", opt);
}

// 管理组（模态框）
async function openPortForwardModal(id) {
    document.getElementById('portForwardForm').reset();

    try {

        // 协议列表
        const protocols = ['tcp4', 'tcp6', 'udp4', 'udp6'];
        const protocolsSelect = document.getElementById('port_forward_protocols');
        protocolsSelect.innerHTML = '';
        protocols.forEach(d => {
            const opt = document.createElement('option');
            opt.value = d.toLowerCase();// 小写
            opt.text = `${d.toUpperCase()}`;// 大写
            protocolsSelect.appendChild(opt);
        });

        // 设备列表
        const { data: devices } = await API.get(`/api/devices`);
        console.log("devices", devices);
        if (devices) {
            const deviceSelect = document.getElementById('port_forward_device_id');
            deviceSelect.innerHTML = '';
            addPlaceholder(deviceSelect, '请选择设备');
            devices.forEach(d => {
                const opt = document.createElement('option');
                opt.value = d.id;
                opt.text = `${d.name}`;
                deviceSelect.appendChild(opt);
            });
        }


        // 编辑
        if (id) {
            const { data: portForward } = await API.get(`/api/port-forwards/${id}`);
            console.log("portForward", portForward);
            if (!portForward) {
                alert('端口转发数据异常');
                return;
            }

            document.getElementById('port_forward_id').value = portForward.id;
            document.getElementById('port_forward_id').disabled = true;
            document.getElementById('port_forward_name').value = portForward.name;
            // document.getElementById('port_forward_description').value = portForward.description || '';
            document.getElementById('port_forward_local_port').value = portForward.local_port;
            document.getElementById('port_forward_remote_host').value = portForward.remote_host;
            document.getElementById('port_forward_remote_port').value = portForward.remote_port;
            document.getElementById('port_forward_device_id').value = portForward.device_id;

            // 高亮已选
            if (portForward.protocols && portForward.protocols.length > 0) {
                const protocolsSelect = document.getElementById('port_forward_protocols');
                Array.from(protocolsSelect.options).forEach(opt => opt.selected = portForward.protocols.includes(opt.value));
            }

            openModal('portForwardModal', 'portForwardModalTitle', '编辑 端口转发');
        }

        // 创建
        if (!id) {
            document.getElementById('port_forward_id').value = '';
            document.getElementById('port_forward_id').disabled = true;

            openModal('portForwardModal', 'portForwardModalTitle', '添加 端口转发');
        }
    } catch (error) {
        alert(`加载IP组信息失败: ${error.message}`);
        console.error("打开IP组模态框失败", error);
    }
}

// 管理组（模态框） -> 表单提交
document.getElementById('portForwardForm').onsubmit = async (e) => {
    e.preventDefault();

    // 表单数据
    const form_id = Number(document.getElementById('port_forward_id').value.trim());
    const form_data = {
        name: document.getElementById('port_forward_name').value.trim(),
        // description: document.getElementById('port_forward_description').value.trim(),
        local_port: Number(document.getElementById('port_forward_local_port').value.trim()),
        remote_host: document.getElementById('port_forward_remote_host').value.trim(),
        remote_port: Number(document.getElementById('port_forward_remote_port').value.trim()),
        protocols: Array.from(document.getElementById('port_forward_protocols').selectedOptions).map(opt => opt.value.trim()),
        device_id: Number(document.getElementById('port_forward_device_id').value.trim()),
    };
    console.log("form_id, form_data", form_id, form_data);

    // 校验
    if (!form_data.name) {
        return alert('请输入端口转发名')
    };
    if (!form_data.local_port) {
        return alert('请输入本地端口')
    };
    if (!form_data.remote_host) {
        return alert('请输入远程主机')
    };
    if (!form_data.remote_port) {
        return alert('请输入远程端口')
    };
    if (form_data.protocols.length === 0) {
        return alert('请选择协议')
    };
    if (!form_data.device_id) {
        return alert('请选择设备')
    };

    try {
        // 提交
        if (isNaN(form_id) || form_id <= 0) {
            await API.post('/api/port-forwards', form_data);
        } else {
            await API.patch(`/api/port-forwards/${form_id}`, form_data);
        }

        // 刷新页面
        location.reload();
    } catch (error) {
        alert(`操作失败: ${error.message}`);
        console.error("端口转发操作失败", error);
    }
};

// 防火墙
async function togglePortForwardForwardStatus(id, checkbox) {
    try {
        await API.patch(`/api/port-forwards/${id}/forward`, { auto_forward: checkbox.checked });
    } catch (error) {
        alert(`防火墙状态更新失败: ${error.message}`);
        console.error("切换防火墙状态失败", error);
        // 恢复 checkbox 状态
        checkbox.checked = !checkbox.checked;
    }
}

// 启用
async function togglePortForwardActiveStatus(id, checkbox) {
    try {
        await API.patch(`/api/port-forwards/${id}/active`, { is_active: checkbox.checked });
    } catch (error) {
        alert(`启用状态更新失败: ${error.message}`);
        console.error("切换启用状态失败", error);
        // 恢复 checkbox 状态
        checkbox.checked = !checkbox.checked;
    }
}

// 删除
async function deletePortForward(id) {
    if (confirm('确定删除此 端口转发？')) {
        try {
            await API.delete(`/api/port-forwards/${id}`);
            // 刷新页面
            location.reload();
        } catch (error) {
            alert(`删除失败: ${error.message}`);
            console.error("删除端口转发失败", error);
        }
    }
}