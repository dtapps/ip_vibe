-- http_log definition

CREATE TABLE `http_log` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `hostname` text NOT NULL, `method` text NOT NULL, `url` text NOT NULL, `request_headers` json NOT NULL, `request_body` json NULL, `status_code` integer NOT NULL, `response_headers` json NOT NULL, `response_body` json NULL, `elapse_time` integer NOT NULL, `is_error` bool NULL DEFAULT (false), `created_at` datetime NOT NULL);

CREATE INDEX `httplog_hostname` ON `http_log` (`hostname`);
CREATE INDEX `httplog_method` ON `http_log` (`method`);
CREATE INDEX `httplog_url` ON `http_log` (`url`);
CREATE INDEX `httplog_status_code` ON `http_log` (`status_code`);
CREATE INDEX `httplog_created_at` ON `http_log` (`created_at`);


-- operation_log definition

CREATE TABLE `operation_log` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `created_at` datetime NOT NULL, `op_type` text NOT NULL, `obj_id` integer NOT NULL, `obj_type` text NOT NULL, `old_value` text NULL, `new_value` text NULL, `status` text NOT NULL, `message` text NOT NULL);

CREATE INDEX `operationlog_op_type` ON `operation_log` (`op_type`);
CREATE INDEX `operationlog_obj_id` ON `operation_log` (`obj_id`);
CREATE INDEX `operationlog_obj_type` ON `operation_log` (`obj_type`);
CREATE INDEX `operationlog_created_at` ON `operation_log` (`created_at`);


-- resty_log definition

CREATE TABLE `resty_log` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `hostname` text NOT NULL, `method` text NOT NULL, `url` text NOT NULL, `request_headers` json NOT NULL, `request_body` json NULL, `status_code` integer NOT NULL, `response_headers` json NOT NULL, `response_body` json NULL, `elapse_time` integer NOT NULL, `is_error` bool NULL DEFAULT (false), `created_at` datetime NOT NULL);

CREATE INDEX `restylog_hostname` ON `resty_log` (`hostname`);
CREATE INDEX `restylog_method` ON `resty_log` (`method`);
CREATE INDEX `restylog_url` ON `resty_log` (`url`);
CREATE INDEX `restylog_status_code` ON `resty_log` (`status_code`);
CREATE INDEX `restylog_created_at` ON `resty_log` (`created_at`);