const IngressUtil = {
  /**
   * 核心解析方法：将远程数据拆解为表单字段
   * @param {string} hostname - 完整主机名 (如: "www.example.com")
   * @param {string} service - 服务字符串 (如: "http://127.0.0.1:8080")
   * @param {Array} domainList - 账号下的域名列表 [{zone_name: "example.com", zone_id: "..."}]
   */
  parseRemoteData(hostname, service, domainList) {
    // 解析 Hostname 和 Domain
    let sub_domain = hostname;
    let domain_name = "";

    if (hostname && domainList) {
      for (const d of domainList) {
        if (hostname === d.zone_name) {
          sub_domain = "@"; // 根域名
          domain_name = d.zone_name;
          break;
        } else if (hostname.endsWith("." + d.zone_name)) {
          sub_domain = hostname.replace("." + d.zone_name, "");
          domain_name = d.zone_name;
          break;
        }
      }
    }

    // 解析 Service 类型、协议和内容
    let service_type = "http";
    let service_protocol = "://";
    let service_content = service || "";

    if (service) {
      if (service.includes("://")) {
        const parts = service.split("://");
        let rawType = parts[0];
        service_content = parts[1];

        // 转换 unix+tls 为表单支持的 value
        service_type = rawType === "unix+tls" ? "unix_tls" : rawType;
        service_protocol = rawType + "://";
      } else if (service.startsWith("http_status:")) {
        service_type = "http_status";
        service_protocol = ":";
        service_content = service.replace("http_status:", "");
      } else if (service === "bastion") {
        service_type = "bastion";
        service_protocol = "---";
        service_content = "";
      }
    }

    return {
      sub_domain,
      domain_name,
      service_type,
      service_protocol,
      service_content,
    };
  },

  /**
   * 核心拼装方法：将表单字段组合为远程提交格式
   */
  buildSubmitData(formEls, domainList) {
    const sub = formEls.sub_domain.value.trim();
    const zoneId = formEls.domain_name.value;
    const type = formEls.service_type.value;
    const content = formEls.service_content.value.trim();

    // 查找域名文字
    const domainItem = domainList.find((d) => d.zone_id == zoneId);
    const zoneName = domainItem ? domainItem.zone_name : "";

    // 拼装 Hostname
    const finalHostname = sub === "@" || !sub ? zoneName : `${sub}.${zoneName}`;

    // 拼装 Service
    let finalService = "";
    if (type === "bastion") {
      finalService = "bastion";
    } else if (type === "http_status") {
      finalService = `http_status:${content}`;
    } else if (type === "unix_tls") {
      finalService = `unix+tls://${content}`;
    } else {
      finalService = `${type}://${content}`;
    }

    return {
      hostname: finalHostname,
      service: finalService,
      path: formEls.domain_path.value.trim(),
    };
  },
};
