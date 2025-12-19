-- device definition

CREATE TABLE "device" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `code` text NOT NULL, `name` text NOT NULL, `description` text NULL, `ip_v4` text NULL, `ip_v6` text NULL, `last_reported_at` datetime NULL, `last_heartbeat_at` datetime NULL);

CREATE UNIQUE INDEX `device_code_key` ON `device` (`code`);
CREATE INDEX `device_last_reported_at` ON `device` (`last_reported_at`);
CREATE INDEX `device_last_heartbeat_at` ON `device` (`last_heartbeat_at`);
CREATE INDEX `device_created_at` ON `device` (`created_at` DESC);


-- domain_monitor definition

CREATE TABLE "domain_monitor" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `name` text NOT NULL, `domain` text NOT NULL, `check_interval` integer NOT NULL DEFAULT (60), `check_protocol` text NOT NULL, `is_active` bool NOT NULL DEFAULT (false), `current_status` text NOT NULL DEFAULT ('unknown'), `last_success_at` datetime NULL, `last_failure_at` datetime NULL, `description` text NULL);

CREATE INDEX `domainmonitor_domain` ON `domain_monitor` (`domain`);
CREATE UNIQUE INDEX `domainmonitor_domain_check_protocol` ON `domain_monitor` (`domain`, `check_protocol`);


-- ip_group definition

CREATE TABLE "ip_group" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `record_id` text NULL, `name` text NOT NULL, `description` text NULL);

CREATE INDEX `ipgroup_created_at` ON `ip_group` (`created_at` DESC);


-- platform_config definition

CREATE TABLE "platform_config" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `capability` text NOT NULL DEFAULT ('all'), `name` text NOT NULL, `description` text NULL, `platform` text NOT NULL, `config_json` json NULL);

CREATE INDEX `platformconfig_capability` ON `platform_config` (`capability`);
CREATE INDEX `platformconfig_platform` ON `platform_config` (`platform`);
CREATE INDEX `platformconfig_created_at` ON `platform_config` (`created_at` DESC);


-- webhook definition

CREATE TABLE "webhook" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `name` text NOT NULL, `description` text NULL, `platform` text NOT NULL, `url` text NOT NULL, `secret` text NULL, `events` json NOT NULL, `template_title` text NULL, `template_content` text NULL, `is_active` bool NOT NULL DEFAULT (false));

CREATE INDEX `webhook_platform` ON `webhook` (`platform`);
CREATE INDEX `webhook_is_active` ON `webhook` (`is_active`);
CREATE INDEX `webhook_created_at` ON `webhook` (`created_at` DESC);


-- cdn_record definition

CREATE TABLE "cdn_record" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `record_id` text NULL, `sub_domain` text NOT NULL, `type` text NOT NULL, `origin_protocol` text NOT NULL DEFAULT ('follow'), `http_port` integer NOT NULL DEFAULT (80), `https_port` integer NOT NULL DEFAULT (443), `cdn_domain` text NULL, `is_active` bool NOT NULL DEFAULT (false), `last_refresh_at` datetime NULL, `ip_group_id` integer NOT NULL, `config_id` integer NOT NULL, CONSTRAINT `cdn_record_ip_group_cdn_records` FOREIGN KEY (`ip_group_id`) REFERENCES `ip_group` (`id`) ON DELETE NO ACTION, CONSTRAINT `cdn_record_platform_config_cdn_records` FOREIGN KEY (`config_id`) REFERENCES `platform_config` (`id`) ON DELETE CASCADE);

CREATE INDEX `cdnrecord_sub_domain` ON `cdn_record` (`sub_domain`);
CREATE INDEX `cdnrecord_type` ON `cdn_record` (`type`);
CREATE INDEX `cdnrecord_is_active` ON `cdn_record` (`is_active`);
CREATE INDEX `cdnrecord_created_at` ON `cdn_record` (`created_at` DESC);
CREATE INDEX `cdnrecord_config_id` ON `cdn_record` (`config_id`);
CREATE INDEX `cdnrecord_ip_group_id` ON `cdn_record` (`ip_group_id`);


-- device_ip_report definition

CREATE TABLE "device_ip_report" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `ip_v4` text NULL, `ip_v6` text NULL, `device_id` integer NOT NULL, CONSTRAINT `device_ip_report_device_ip_reports` FOREIGN KEY (`device_id`) REFERENCES `device` (`id`) ON DELETE CASCADE);

CREATE INDEX `deviceipreport_created_at` ON `device_ip_report` (`created_at` DESC);
CREATE INDEX `deviceipreport_device_id` ON `device_ip_report` (`device_id`);


-- device_port_forward definition

CREATE TABLE "device_port_forward" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `name` text NOT NULL, `description` text NULL, `local_port` integer NOT NULL, `remote_host` text NOT NULL, `remote_port` integer NOT NULL, `protocols` json NOT NULL, `auto_forward` bool NOT NULL DEFAULT (false), `is_active` bool NOT NULL DEFAULT (false), `device_port_forwards` integer NOT NULL, CONSTRAINT `device_port_forward_device_port_forwards` FOREIGN KEY (`device_port_forwards`) REFERENCES `device` (`id`) ON DELETE CASCADE);

CREATE INDEX `deviceportforward_is_active` ON `device_port_forward` (`is_active`);
CREATE INDEX `deviceportforward_created_at` ON `device_port_forward` (`created_at` DESC);


-- dns_record definition

CREATE TABLE "dns_record" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `record_id` text NULL, `sub_domain` text NOT NULL, `type` text NOT NULL, `is_active` bool NOT NULL DEFAULT (false), `last_refresh_at` datetime NULL, `cdn_record_id` integer NULL, `ip_group_id` integer NULL, `config_id` integer NOT NULL, CONSTRAINT `dns_record_cdn_record_dns_records` FOREIGN KEY (`cdn_record_id`) REFERENCES `cdn_record` (`id`) ON DELETE SET NULL, CONSTRAINT `dns_record_ip_group_dns_records` FOREIGN KEY (`ip_group_id`) REFERENCES `ip_group` (`id`) ON DELETE SET NULL, CONSTRAINT `dns_record_platform_config_dns_records` FOREIGN KEY (`config_id`) REFERENCES `platform_config` (`id`) ON DELETE CASCADE);

CREATE INDEX `dnsrecord_sub_domain` ON `dns_record` (`sub_domain`);
CREATE INDEX `dnsrecord_type` ON `dns_record` (`type`);
CREATE INDEX `dnsrecord_is_active` ON `dns_record` (`is_active`);
CREATE INDEX `dnsrecord_created_at` ON `dns_record` (`created_at` DESC);
CREATE INDEX `dnsrecord_config_id` ON `dns_record` (`config_id`);
CREATE INDEX `dnsrecord_ip_group_id` ON `dns_record` (`ip_group_id`);
CREATE INDEX `dnsrecord_cdn_record_id` ON `dns_record` (`cdn_record_id`);


-- domain_monitor_history definition

CREATE TABLE `domain_monitor_history` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `check_time` datetime NOT NULL, `is_success` bool NOT NULL, `check_duration` integer NULL, `status_message` text NULL, `http_status` integer NULL, `ip_resolved` json NULL, `error_log` text NULL, `ssl_expiry_date` datetime NULL, `ssl_common_name` text NULL, `ssl_issuer_name` text NULL, `ssl_signature_alg` text NULL, `ssl_san_list` text NULL, `ssl_certificate_chain` text NULL, `monitor_id` integer NOT NULL, CONSTRAINT `domain_monitor_history_domain_monitor_history_records` FOREIGN KEY (`monitor_id`) REFERENCES `domain_monitor` (`id`) ON DELETE CASCADE);

CREATE INDEX `domainmonitorhistory_monitor_id_check_time` ON `domain_monitor_history` (`monitor_id`, `check_time`);
CREATE INDEX `domainmonitorhistory_check_time` ON `domain_monitor_history` (`check_time` DESC);


-- ip_group_device definition

CREATE TABLE "ip_group_device" (`created_at` datetime NOT NULL, `device_id` integer NOT NULL, `ip_group_id` integer NOT NULL, PRIMARY KEY (`device_id`, `ip_group_id`), CONSTRAINT `ip_group_device_device_device` FOREIGN KEY (`device_id`) REFERENCES `device` (`id`) ON DELETE NO ACTION, CONSTRAINT `ip_group_device_ip_group_ip_group` FOREIGN KEY (`ip_group_id`) REFERENCES `ip_group` (`id`) ON DELETE NO ACTION);

CREATE INDEX `ipgroupdevice_created_at` ON `ip_group_device` (`created_at` DESC);
CREATE INDEX `ipgroupdevice_device_id` ON `ip_group_device` (`device_id`);
CREATE INDEX `ipgroupdevice_ip_group_id` ON `ip_group_device` (`ip_group_id`);


-- webhook_subscription definition

CREATE TABLE "webhook_subscription" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `name` text NOT NULL, `description` text NULL, `obj_type` text NOT NULL, `obj_id` integer NOT NULL, `is_active` bool NOT NULL DEFAULT (false), `webhook_subscriptions` integer NULL, `webhook_id` integer NOT NULL, CONSTRAINT `webhook_subscription_webhook_subscriptions` FOREIGN KEY (`webhook_subscriptions`) REFERENCES `webhook` (`id`) ON DELETE CASCADE, CONSTRAINT `webhook_subscription_webhook_webhook` FOREIGN KEY (`webhook_id`) REFERENCES `webhook` (`id`) ON DELETE NO ACTION);

CREATE UNIQUE INDEX `webhooksubscription_webhook_id_obj_type_obj_id` ON `webhook_subscription` (`webhook_id`, `obj_type`, `obj_id`);
CREATE INDEX `webhooksubscription_created_at` ON `webhook_subscription` (`created_at` DESC);
CREATE INDEX `webhooksubscription_webhook_id` ON `webhook_subscription` (`webhook_id`);