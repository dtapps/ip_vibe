-- cache definition

CREATE TABLE `cache` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `key` text NOT NULL, `value` blob NULL);

CREATE UNIQUE INDEX `cache_key_key` ON `cache` (`key`);
CREATE UNIQUE INDEX `cache_key` ON `cache` (`key`);


-- device definition

CREATE TABLE "device" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `code` text NOT NULL, `name` text NOT NULL, `description` text NULL, `ip_v4` text NULL, `ip_v6` text NULL, `last_reported_at` datetime NULL, `last_heartbeat_at` datetime NULL);

CREATE UNIQUE INDEX `device_code_key` ON `device` (`code`);
CREATE INDEX `device_last_reported_at` ON `device` (`last_reported_at`);
CREATE INDEX `device_last_heartbeat_at` ON `device` (`last_heartbeat_at`);
CREATE INDEX `device_created_at` ON `device` (`created_at` DESC);


-- device_group definition

CREATE TABLE `device_group` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `code` text NOT NULL, `name` text NOT NULL, `description` text NULL);

CREATE INDEX `devicegroup_created_at` ON `device_group` (`created_at` DESC);
CREATE UNIQUE INDEX `devicegroup_code` ON `device_group` (`code`);


-- domain_monitor definition

CREATE TABLE "domain_monitor" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `name` text NOT NULL, `domain` text NOT NULL, `check_interval` integer NOT NULL DEFAULT (60), `check_protocol` text NOT NULL, `is_active` bool NOT NULL DEFAULT (false), `current_status` text NOT NULL DEFAULT ('unknown'), `last_success_at` datetime NULL, `last_failure_at` datetime NULL, `description` text NULL, `request_headers` json NULL);

CREATE INDEX `domainmonitor_domain` ON `domain_monitor` (`domain`);
CREATE UNIQUE INDEX `domainmonitor_domain_check_protocol` ON `domain_monitor` (`domain`, `check_protocol`);
CREATE INDEX `domainmonitor_created_at` ON `domain_monitor` (`created_at` DESC);


-- ip_group definition

CREATE TABLE ip_group (
	id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL,
	code TEXT NOT NULL,
	record_id TEXT,
	name TEXT NOT NULL,
	description TEXT
);

CREATE INDEX `ipgroup_created_at` ON `ip_group` (`created_at` DESC);
CREATE UNIQUE INDEX `ipgroup_code` ON `ip_group` (`code`);


-- platform_config definition

CREATE TABLE "platform_config" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `capability` text NOT NULL DEFAULT ('all'), `name` text NOT NULL, `description` text NULL, `platform` text NOT NULL, `config_json` json NULL);

CREATE INDEX `platformconfig_capability` ON `platform_config` (`capability`);
CREATE INDEX `platformconfig_platform` ON `platform_config` (`platform`);
CREATE INDEX `platformconfig_created_at` ON `platform_config` (`created_at` DESC);


-- ssl_import definition

CREATE TABLE "ssl_import" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `source` text NOT NULL DEFAULT ('unknown'), `domain` text NOT NULL, `serial_number` text NOT NULL, `certificate` text NOT NULL, `private_key` text NOT NULL, `not_before` datetime NOT NULL, `not_after` datetime NOT NULL, `status` text NOT NULL DEFAULT ('unknown'));

CREATE UNIQUE INDEX `ssl_import_serial_number_key` ON `ssl_import` (`serial_number`);
CREATE INDEX `sslimport_created_at` ON `ssl_import` (`created_at` DESC);
CREATE INDEX `sslimport_domain` ON `ssl_import` (`domain`);
CREATE UNIQUE INDEX `sslimport_serial_number` ON `ssl_import` (`serial_number`);


-- webhook definition

CREATE TABLE "webhook" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `name` text NOT NULL, `description` text NULL, `platform` text NOT NULL, `url` text NOT NULL, `secret` text NULL, `events` json NOT NULL, `template_title` text NULL, `template_content` text NULL, `is_active` bool NOT NULL DEFAULT (false));

CREATE INDEX `webhook_platform` ON `webhook` (`platform`);
CREATE INDEX `webhook_is_active` ON `webhook` (`is_active`);
CREATE INDEX `webhook_created_at` ON `webhook` (`created_at` DESC);


-- cdn_record definition

CREATE TABLE "cdn_record" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `record_id` text NULL, `sub_domain` text NOT NULL, `type` text NOT NULL, `config_json` json NULL, `cdn_domain` text NULL, `is_active` bool NOT NULL DEFAULT (false), `last_refresh_at` datetime NULL, `last_api_response` text NULL, `last_api_response_at` datetime NULL, `device_group_id` integer NULL, `ip_group_id` integer NULL, `config_id` integer NOT NULL, CONSTRAINT `cdn_record_device_group_cdn_records` FOREIGN KEY (`device_group_id`) REFERENCES `device_group` (`id`) ON DELETE SET NULL, CONSTRAINT `cdn_record_ip_group_cdn_records` FOREIGN KEY (`ip_group_id`) REFERENCES `ip_group` (`id`) ON DELETE SET NULL, CONSTRAINT `cdn_record_platform_config_cdn_records` FOREIGN KEY (`config_id`) REFERENCES `platform_config` (`id`) ON DELETE CASCADE);

CREATE INDEX `cdnrecord_created_at` ON `cdn_record` (`created_at` DESC);
CREATE INDEX `cdnrecord_config_id` ON `cdn_record` (`config_id`);
CREATE INDEX `cdnrecord_record_id` ON `cdn_record` (`record_id`);
CREATE INDEX `cdnrecord_sub_domain` ON `cdn_record` (`sub_domain`);
CREATE INDEX `cdnrecord_type` ON `cdn_record` (`type`);
CREATE INDEX `cdnrecord_ip_group_id` ON `cdn_record` (`ip_group_id`);
CREATE INDEX `cdnrecord_is_active` ON `cdn_record` (`is_active`);
CREATE INDEX `cdnrecord_device_group_id` ON `cdn_record` (`device_group_id`);


-- device_group_device definition

CREATE TABLE `device_group_device` (`created_at` datetime NOT NULL, `device_id` integer NOT NULL, `device_group_id` integer NOT NULL, PRIMARY KEY (`device_id`, `device_group_id`), CONSTRAINT `device_group_device_device_device` FOREIGN KEY (`device_id`) REFERENCES `device` (`id`) ON DELETE NO ACTION, CONSTRAINT `device_group_device_device_group_device_group` FOREIGN KEY (`device_group_id`) REFERENCES `device_group` (`id`) ON DELETE NO ACTION);

CREATE INDEX `devicegroupdevice_created_at` ON `device_group_device` (`created_at` DESC);
CREATE INDEX `devicegroupdevice_device_id` ON `device_group_device` (`device_id`);
CREATE INDEX `devicegroupdevice_device_group_id` ON `device_group_device` (`device_group_id`);


-- device_port_forward definition

CREATE TABLE "device_port_forward" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `name` text NOT NULL, `description` text NULL, `local_port` integer NOT NULL, `remote_host` text NOT NULL, `remote_port` integer NOT NULL, `protocols` json NOT NULL, `auto_forward` bool NOT NULL DEFAULT (false), `is_active` bool NOT NULL DEFAULT (false), `device_port_forwards` integer NOT NULL, CONSTRAINT `device_port_forward_device_port_forwards` FOREIGN KEY (`device_port_forwards`) REFERENCES `device` (`id`) ON DELETE CASCADE);

CREATE INDEX `deviceportforward_is_active` ON `device_port_forward` (`is_active`);
CREATE INDEX `deviceportforward_created_at` ON `device_port_forward` (`created_at` DESC);


-- dns_record definition

CREATE TABLE "dns_record" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `record_id` text NULL, `sub_domain` text NOT NULL, `type` text NOT NULL, `config_json` json NULL, `is_active` bool NOT NULL DEFAULT (false), `last_refresh_at` datetime NULL, `last_api_response` text NULL, `last_api_response_at` datetime NULL, `cdn_record_id` integer NULL, `device_group_id` integer NULL, `ip_group_id` integer NULL, `config_id` integer NOT NULL, CONSTRAINT `dns_record_cdn_record_dns_records` FOREIGN KEY (`cdn_record_id`) REFERENCES `cdn_record` (`id`) ON DELETE SET NULL, CONSTRAINT `dns_record_device_group_dns_records` FOREIGN KEY (`device_group_id`) REFERENCES `device_group` (`id`) ON DELETE SET NULL, CONSTRAINT `dns_record_ip_group_dns_records` FOREIGN KEY (`ip_group_id`) REFERENCES `ip_group` (`id`) ON DELETE SET NULL, CONSTRAINT `dns_record_platform_config_dns_records` FOREIGN KEY (`config_id`) REFERENCES `platform_config` (`id`) ON DELETE CASCADE);

CREATE INDEX `dnsrecord_created_at` ON `dns_record` (`created_at` DESC);
CREATE INDEX `dnsrecord_config_id` ON `dns_record` (`config_id`);
CREATE INDEX `dnsrecord_record_id` ON `dns_record` (`record_id`);
CREATE INDEX `dnsrecord_sub_domain` ON `dns_record` (`sub_domain`);
CREATE INDEX `dnsrecord_type` ON `dns_record` (`type`);
CREATE INDEX `dnsrecord_ip_group_id` ON `dns_record` (`ip_group_id`);
CREATE INDEX `dnsrecord_device_group_id` ON `dns_record` (`device_group_id`);
CREATE INDEX `dnsrecord_cdn_record_id` ON `dns_record` (`cdn_record_id`);
CREATE INDEX `dnsrecord_is_active` ON `dns_record` (`is_active`);


-- ip_group_device definition

CREATE TABLE "ip_group_device" (`created_at` datetime NOT NULL, `device_id` integer NOT NULL, `ip_group_id` integer NOT NULL, PRIMARY KEY (`device_id`, `ip_group_id`), CONSTRAINT `ip_group_device_device_device` FOREIGN KEY (`device_id`) REFERENCES `device` (`id`) ON DELETE NO ACTION, CONSTRAINT `ip_group_device_ip_group_ip_group` FOREIGN KEY (`ip_group_id`) REFERENCES `ip_group` (`id`) ON DELETE NO ACTION);

CREATE INDEX `ipgroupdevice_created_at` ON `ip_group_device` (`created_at` DESC);
CREATE INDEX `ipgroupdevice_device_id` ON `ip_group_device` (`device_id`);
CREATE INDEX `ipgroupdevice_ip_group_id` ON `ip_group_device` (`ip_group_id`);


-- ssl_record definition

CREATE TABLE "ssl_record" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `domain` text NOT NULL, `cert_domain` text NULL, `cert_url` text NULL, `cert_stable_url` text NULL, `cert_private_key` text NULL, `cert_certificate` text NULL, `cert_issuer_certificate` text NULL, `cert_csr` text NULL, `not_before` datetime NULL, `not_after` datetime NULL, `status` text NOT NULL DEFAULT ('unknown'), `provider` text NOT NULL DEFAULT ('unknown'), `is_active` bool NOT NULL DEFAULT (false), `renewal_days` integer NULL DEFAULT (30), `is_auto_renewal` bool NULL DEFAULT (false), `last_refresh_at` datetime NULL, `last_api_response` text NULL, `last_api_response_at` datetime NULL, `config_id` integer NOT NULL, CONSTRAINT `ssl_record_platform_config_ssl_records` FOREIGN KEY (`config_id`) REFERENCES `platform_config` (`id`) ON DELETE CASCADE);

CREATE INDEX `sslrecord_created_at` ON `ssl_record` (`created_at` DESC);
CREATE INDEX `sslrecord_domain` ON `ssl_record` (`domain`);
CREATE INDEX `sslrecord_not_before` ON `ssl_record` (`not_before`);
CREATE INDEX `sslrecord_not_after` ON `ssl_record` (`not_after` DESC);
CREATE INDEX `sslrecord_status` ON `ssl_record` (`status`);
CREATE INDEX `sslrecord_provider` ON `ssl_record` (`provider`);
CREATE INDEX `sslrecord_is_active` ON `ssl_record` (`is_active`);
CREATE INDEX `sslrecord_is_auto_renewal` ON `ssl_record` (`is_auto_renewal`);


-- webhook_subscription definition

CREATE TABLE "webhook_subscription" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `name` text NOT NULL, `description` text NULL, `obj_type` text NOT NULL, `obj_id` integer NOT NULL, `is_active` bool NOT NULL DEFAULT (false), `webhook_subscriptions` integer NULL, `webhook_id` integer NOT NULL, CONSTRAINT `webhook_subscription_webhook_subscriptions` FOREIGN KEY (`webhook_subscriptions`) REFERENCES `webhook` (`id`) ON DELETE CASCADE, CONSTRAINT `webhook_subscription_webhook_webhook` FOREIGN KEY (`webhook_id`) REFERENCES `webhook` (`id`) ON DELETE NO ACTION);

CREATE INDEX `webhooksubscription_created_at` ON `webhook_subscription` (`created_at` DESC);
CREATE INDEX `webhooksubscription_webhook_id` ON `webhook_subscription` (`webhook_id`);
CREATE INDEX `webhooksubscription_obj_type` ON `webhook_subscription` (`obj_type`);
CREATE INDEX `webhooksubscription_obj_id` ON `webhook_subscription` (`obj_id`);
CREATE INDEX `webhooksubscription_is_active` ON `webhook_subscription` (`is_active`);
CREATE INDEX `webhooksubscription_obj_type_obj_id` ON `webhook_subscription` (`obj_type`, `obj_id`);
CREATE INDEX `webhooksubscription_obj_type_obj_id_created_at` ON `webhook_subscription` (`obj_type`, `obj_id`, `created_at`);
CREATE UNIQUE INDEX `webhooksubscription_webhook_id_obj_type_obj_id` ON `webhook_subscription` (`webhook_id`, `obj_type`, `obj_id`);