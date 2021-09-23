#pragma once

#ifdef __cplusplus
extern "C" {
#endif

typedef struct {
  const char* key;
  const char* value;
} traceable_attribute;

typedef struct {
  int count;
  const traceable_attribute* attribute_array;
} traceable_attributes;

typedef enum { TRACEABLE_LOG_NONE, TRACEABLE_LOG_STDOUT } TRACEABLE_LOG_MODE;

typedef struct {
  TRACEABLE_LOG_MODE mode;
} traceable_log_configuration;

typedef struct {
  const char* opa_server_url;
  const char* logging_dir;
  const char* logging_file_prefix;
  int log_to_console;
  int skip_verify;
  int min_delay;
  int max_delay;
  int debug_log;
} traceable_opa_config;

typedef struct {
  int enabled;
} traceable_modsecurity_config;

typedef struct {
  int enabled;
} traceable_rangeblocking_config;

typedef struct {
  int enabled;
  const char* remote_endpoint;
  int poll_period_sec;
} traceable_remote_config;

typedef struct {
  traceable_log_configuration log_config;
  traceable_opa_config opa_config;
  traceable_modsecurity_config modsecurity_config;
  traceable_rangeblocking_config rb_config;
  traceable_remote_config remote_config;
  int evaluate_body;
} traceable_blocking_config;

typedef struct {
  int block;
  traceable_attributes attributes;
} traceable_block_result;

typedef enum { TRACEABLE_SUCCESS, TRACEABLE_FAIL } TRACEABLE_RET;

typedef void* traceable_blocking_engine;

TRACEABLE_RET traceable_new_blocking_engine(
    traceable_blocking_config blocking_config,
    traceable_blocking_engine* out_blocking_engine);

TRACEABLE_RET traceable_start_blocking_engine(
    traceable_blocking_engine blocking_engine);

TRACEABLE_RET traceable_block_request(traceable_blocking_engine blocking_engine,
                                      traceable_attributes attributes,
                                      traceable_block_result* out_block_result);

TRACEABLE_RET traceable_block_request_headers(
    traceable_blocking_engine blocking_engine, traceable_attributes attributes,
    traceable_block_result* out_block_result);

TRACEABLE_RET traceable_block_request_body(
    traceable_blocking_engine blocking_engine, traceable_attributes attributes,
    traceable_block_result* out_block_result);

TRACEABLE_RET traceable_delete_block_result_data(traceable_block_result result);

TRACEABLE_RET traceable_delete_blocking_engine(
    traceable_blocking_engine blocking_engine);

#ifdef __cplusplus
}
#endif
