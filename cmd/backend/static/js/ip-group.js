// 选择设备 devices
function onIpGroupDevicesSelect(sel) {
    const opt = sel.options[sel.selectedIndex];
    // console.log("onIpGroupDevicesSelect", opt);
}

// 管理组（模态框）
async function openIpGroupModal(id) {
    document.getElementById('ipGroupForm').reset();

    try {
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

            openModal('ipGroupModal', 'ipGroupModalTitle', '编辑 IP 组');
        }

        // 创建
        if (!id) {
            document.getElementById('ip_group_id').value = '';
            document.getElementById('ip_group_id').disabled = true;

            openModal('ipGroupModal', 'ipGroupModalTitle', '添加 IP 组');
        }
    } catch (error) {
        alert(`加载IP组信息失败: ${error.message}`);
        console.error("打开IP组模态框失败", error);
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

    try {
        // 提交
        if (isNaN(form_id) || form_id <= 0) {
            await API.post('/api/ip-groups', form_data);
        } else {
            await API.patch(`/api/ip-groups/${form_id}`, form_data);
        }

        // 刷新页面
        location.reload();
    } catch (error) {
        alert(`操作失败: ${error.message}`);
        console.error("IP组操作失败", error);
    }
};

// 删除
async function deleteIpGroup(id) {
    if (confirm('确定删除此 IP 组？')) {
        try {
            await API.delete(`/api/ip-groups/${id}`);
            // 刷新页面
            location.reload();
        } catch (error) {
            alert(`删除失败: ${error.message}`);
            console.error("删除IP组失败", error);
        }
    }
}

// 实时 设备 IP 地址 弹窗
async function viewLiveDeviceIpAddresse(id, name) {
    try {
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
    } catch (error) {
        alert(`加载IP组设备信息失败: ${error.message}`);
        console.error("查看设备IP地址失败", error);
    }
}
