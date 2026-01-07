-- device_ip_report definition

CREATE TABLE `device_ip_report` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `device_id` integer NOT NULL, `ip_v4` text NULL, `ip_v6` text NULL);

CREATE INDEX `deviceipreport_created_at` ON `device_ip_report` (`created_at` DESC);
CREATE INDEX `deviceipreport_device_id` ON `device_ip_report` (`device_id`);
CREATE INDEX `deviceipreport_device_id_created_at` ON `device_ip_report` (`device_id`, `created_at`);


-- domain_monitor_history definition

CREATE TABLE `domain_monitor_history` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `monitor_id` integer NOT NULL, `check_time` datetime NOT NULL, `is_success` bool NOT NULL, `check_duration` integer NULL, `status_message` text NULL, `http_status` integer NULL, `ip_resolved` json NULL, `error_log` text NULL, `ssl_expiry_date` datetime NULL, `ssl_common_name` text NULL, `ssl_issuer_name` text NULL, `ssl_signature_alg` text NULL, `ssl_san_list` text NULL, `ssl_certificate_chain` text NULL, `ssl_not_date` datetime NULL, `ssl_serial_number` text NULL, `ssl_version` integer NULL);

CREATE INDEX `domainmonitorhistory_created_at` ON `domain_monitor_history` (`created_at` DESC);
CREATE INDEX `domainmonitorhistory_monitor_id` ON `domain_monitor_history` (`monitor_id`);
CREATE INDEX `domainmonitorhistory_check_time` ON `domain_monitor_history` (`check_time` DESC);
CREATE INDEX `domainmonitorhistory_monitor_id_check_time` ON `domain_monitor_history` (`monitor_id`, `check_time`);


-- http_log definition

CREATE TABLE "http_log" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `hostname` text NOT NULL, `method` text NOT NULL, `url` text NOT NULL, `request_headers` json NULL, `request_body` json NULL, `status_code` integer NOT NULL, `response_headers` json NULL, `response_body` json NULL, `elapse_time` integer NOT NULL, `is_error` bool NULL DEFAULT (false), `created_at` datetime NOT NULL, `go_version` text NULL DEFAULT ('go1.26rc1'), `plugin_version` text NULL DEFAULT ('1.0.2'));

CREATE INDEX `httplog_hostname` ON `http_log` (`hostname`);
CREATE INDEX `httplog_method` ON `http_log` (`method`);
CREATE INDEX `httplog_url` ON `http_log` (`url`);
CREATE INDEX `httplog_status_code` ON `http_log` (`status_code`);
CREATE INDEX `httplog_created_at` ON `http_log` (`created_at` DESC);


-- operation_log definition

CREATE TABLE "operation_log" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `source` text NOT NULL DEFAULT ('unknown'), `op_type` text NOT NULL, `obj_type` text NOT NULL, `obj_id` integer NOT NULL, `old_value` json NULL, `new_value` json NULL, `status` text NOT NULL, `message` text NOT NULL);

CREATE INDEX `operationlog_created_at` ON `operation_log` (`created_at` DESC);
CREATE INDEX `operationlog_op_type` ON `operation_log` (`op_type`);
CREATE INDEX `operationlog_obj_type` ON `operation_log` (`obj_type`);
CREATE INDEX `operationlog_obj_id` ON `operation_log` (`obj_id`);
CREATE INDEX `operationlog_obj_type_obj_id` ON `operation_log` (`obj_type`, `obj_id`);
CREATE INDEX `operationlog_obj_type_obj_id_created_at` ON `operation_log` (`obj_type`, `obj_id`, `created_at`);


-- port_forward_stats definition

CREATE TABLE `port_forward_stats` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `device_id` integer NOT NULL, `port_forward_id` integer NOT NULL, `start_time` datetime NOT NULL, `active_conn` integer NULL, `total_conn` integer NULL, `bytes_in` integer NULL, `bytes_out` integer NULL, `udp_packets_in` integer NULL, `udp_packets_out` integer NULL, `errors` integer NULL);

CREATE INDEX `portforwardstats_created_at` ON `port_forward_stats` (`created_at` DESC);
CREATE INDEX `portforwardstats_device_id` ON `port_forward_stats` (`device_id`);
CREATE INDEX `portforwardstats_port_forward_id` ON `port_forward_stats` (`port_forward_id`);
CREATE INDEX `portforwardstats_device_id_created_at` ON `port_forward_stats` (`device_id`, `created_at`);
CREATE INDEX `portforwardstats_port_forward_id_created_at` ON `port_forward_stats` (`port_forward_id`, `created_at`);
CREATE INDEX `portforwardstats_device_id_port_forward_id_created_at` ON `port_forward_stats` (`device_id`, `port_forward_id`, `created_at`);


-- resty_log definition

CREATE TABLE "resty_log" (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `hostname` text NOT NULL, `method` text NOT NULL, `url` text NOT NULL, `request_headers` json NULL, `request_body` json NULL, `status_code` integer NOT NULL, `response_headers` json NULL, `response_body` json NULL, `elapse_time` integer NOT NULL, `is_error` bool NULL DEFAULT (false), `created_at` datetime NOT NULL, `go_version` text NULL DEFAULT ('go1.26rc1'), `plugin_version` text NULL DEFAULT ('1.0.11'), `resty_version` text NULL DEFAULT ('3.0.0-beta.6'));

CREATE INDEX `restylog_hostname` ON `resty_log` (`hostname`);
CREATE INDEX `restylog_method` ON `resty_log` (`method`);
CREATE INDEX `restylog_url` ON `resty_log` (`url`);
CREATE INDEX `restylog_status_code` ON `resty_log` (`status_code`);
CREATE INDEX `restylog_created_at` ON `resty_log` (`created_at` DESC);


-- ssl_certificate_history definition

CREATE TABLE `ssl_certificate_history` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `ssl_record_id` integer NOT NULL, `cert_domain` text NULL, `cert_url` text NULL, `cert_stable_url` text NULL, `cert_private_key` text NOT NULL, `cert_certificate` text NOT NULL, `cert_issuer_certificate` text NULL, `cert_csr` text NULL, `not_before` datetime NULL, `not_after` datetime NULL, `provider` text NOT NULL DEFAULT ('letsencrypt'));

CREATE INDEX `sslcertificatehistory_created_at` ON `ssl_certificate_history` (`created_at` DESC);
CREATE INDEX `sslcertificatehistory_ssl_record_id` ON `ssl_certificate_history` (`ssl_record_id`);


-- webhook_history definition

CREATE TABLE `webhook_history` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `webhook_id` integer NOT NULL, `subscription_id` integer NULL, `url` text NOT NULL, `title` text NULL, `content` text NULL, `response` json NULL, `success` bool NOT NULL DEFAULT (false), `error_message` text NULL);

CREATE INDEX `webhookhistory_created_at` ON `webhook_history` (`created_at` DESC);
CREATE INDEX `webhookhistory_webhook_id` ON `webhook_history` (`webhook_id`);
CREATE INDEX `webhookhistory_subscription_id` ON `webhook_history` (`subscription_id`);
CREATE INDEX `webhookhistory_success` ON `webhook_history` (`success`);
CREATE INDEX `webhookhistory_webhook_id_created_at` ON `webhook_history` (`webhook_id`, `created_at`);
CREATE INDEX `webhookhistory_success_created_at` ON `webhook_history` (`success`, `created_at`);