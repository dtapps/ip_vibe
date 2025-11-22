-- device definition

CREATE TABLE `device` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `code` text NOT NULL, `name` text NOT NULL, `ip_v4` text NULL, `ip_v6` text NULL, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL);

CREATE UNIQUE INDEX `device_code_key` ON `device` (`code`);
CREATE UNIQUE INDEX `device_code` ON `device` (`code`);


-- ip_group definition

CREATE TABLE `ip_group` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `name` text NOT NULL, `description` text NULL, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL);


-- platform_config definition

CREATE TABLE `platform_config` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `name` text NOT NULL, `platform` text NOT NULL, `config_json` json NULL, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL);

CREATE INDEX `platformconfig_platform` ON `platform_config` (`platform`);


-- cdn_record definition

CREATE TABLE `cdn_record` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `record_id` text NULL, `domain` text NOT NULL, `type` text NOT NULL, `cdn_domain` text NULL, `is_active` bool NOT NULL DEFAULT (false), `last_refresh` datetime NULL, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `ip_group_id` integer NOT NULL, `config_id` integer NOT NULL, CONSTRAINT `cdn_record_ip_group_cdn_records` FOREIGN KEY (`ip_group_id`) REFERENCES `ip_group` (`id`) ON DELETE NO ACTION, CONSTRAINT `cdn_record_platform_config_cdn_records` FOREIGN KEY (`config_id`) REFERENCES `platform_config` (`id`) ON DELETE CASCADE);

CREATE INDEX `cdnrecord_config_id` ON `cdn_record` (`config_id`);
CREATE INDEX `cdnrecord_type` ON `cdn_record` (`type`);
CREATE INDEX `cdnrecord_ip_group_id` ON `cdn_record` (`ip_group_id`);
CREATE INDEX `cdnrecord_is_active` ON `cdn_record` (`is_active`);


-- device_ip_report definition

CREATE TABLE `device_ip_report` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `ip_v4` text NULL, `ip_v6` text NULL, `created_at` datetime NOT NULL, `device_id` integer NOT NULL, CONSTRAINT `device_ip_report_device_ip_reports` FOREIGN KEY (`device_id`) REFERENCES `device` (`id`) ON DELETE CASCADE);

CREATE INDEX `deviceipreport_device_id` ON `device_ip_report` (`device_id`);


-- dns_record definition

CREATE TABLE `dns_record` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `record_id` text NULL, `domain` text NOT NULL, `type` text NOT NULL, `is_active` bool NOT NULL DEFAULT (false), `last_refresh` datetime NULL, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `cdn_record_id` integer NULL, `ip_group_id` integer NULL, `config_id` integer NOT NULL, CONSTRAINT `dns_record_cdn_record_dns_records` FOREIGN KEY (`cdn_record_id`) REFERENCES `cdn_record` (`id`) ON DELETE SET NULL, CONSTRAINT `dns_record_ip_group_dns_records` FOREIGN KEY (`ip_group_id`) REFERENCES `ip_group` (`id`) ON DELETE SET NULL, CONSTRAINT `dns_record_platform_config_dns_records` FOREIGN KEY (`config_id`) REFERENCES `platform_config` (`id`) ON DELETE CASCADE);

CREATE INDEX `dnsrecord_config_id` ON `dns_record` (`config_id`);
CREATE INDEX `dnsrecord_type` ON `dns_record` (`type`);
CREATE INDEX `dnsrecord_ip_group_id` ON `dns_record` (`ip_group_id`);
CREATE INDEX `dnsrecord_cdn_record_id` ON `dns_record` (`cdn_record_id`);
CREATE INDEX `dnsrecord_is_active` ON `dns_record` (`is_active`);


-- ip_group_device definition

CREATE TABLE `ip_group_device` (`created_at` datetime NOT NULL, `device_id` integer NOT NULL, `ip_group_id` integer NOT NULL, PRIMARY KEY (`device_id`, `ip_group_id`), CONSTRAINT `ip_group_device_device_device` FOREIGN KEY (`device_id`) REFERENCES `device` (`id`) ON DELETE NO ACTION, CONSTRAINT `ip_group_device_ip_group_ip_group` FOREIGN KEY (`ip_group_id`) REFERENCES `ip_group` (`id`) ON DELETE NO ACTION);


-- ddns_log definition

CREATE TABLE "ddns_log" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `old_value` text NULL, `new_value` text NOT NULL, `status` text NOT NULL, `message` bool NOT NULL DEFAULT (false), `created_at` datetime NOT NULL, `record_id` integer NOT NULL, CONSTRAINT `ddns_log_dns_record_ddns_logs` FOREIGN KEY (`record_id`) REFERENCES `dns_record` (`id`) ON DELETE CASCADE);

CREATE INDEX `ddnslog_record_id` ON `ddns_log` (`record_id`);
CREATE INDEX `ddnslog_status` ON `ddns_log` (`status`);