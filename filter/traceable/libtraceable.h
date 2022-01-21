#ifndef LIBTRACEABLE_H
#define LIBTRACEABLE_H

#ifdef __cplusplus
extern "C" {
#endif

/*
 * Traceable input and output structures
 */

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
  const char* cert_file;
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
  const char* cert_file;
} traceable_remote_config;

typedef struct {
  int enabled;
  traceable_opa_config opa_config;
  traceable_modsecurity_config modsecurity_config;
  traceable_rangeblocking_config rb_config;
  int evaluate_body;
  int skip_internal_request;
} traceable_blocking_config;

typedef struct {
  const char* service_name;
} traceable_agent_config;

typedef struct {
  int block;
  int sample;
  traceable_attributes attributes;
} traceable_process_request_result;

typedef struct {
  traceable_log_configuration log_config;
  traceable_remote_config remote_config;
  traceable_blocking_config blocking_config;
  traceable_agent_config agent_config;
} traceable_libtraceable_config;

typedef struct {
  int sample;
  traceable_attributes attributes;
} traceable_process_span_result;

typedef enum { TRACEABLE_SUCCESS, TRACEABLE_FAIL } TRACEABLE_RET;

typedef void* traceable_libtraceable;

/*
 * Traceable api functions
 */
TRACEABLE_RET traceable_new_libtraceable(
    traceable_libtraceable_config libtraceable_config,
    traceable_libtraceable* out_libtraceable);
TRACEABLE_RET traceable_start_libtraceable(traceable_libtraceable libtraceable);
TRACEABLE_RET traceable_delete_libtraceable(
    traceable_libtraceable libtraceable);

TRACEABLE_RET traceable_process_request_headers(
    traceable_libtraceable libtraceable, traceable_attributes attributes,
    traceable_process_request_result* out_process_result);
TRACEABLE_RET traceable_process_request_body(
    traceable_libtraceable libtraceable, traceable_attributes attributes,
    traceable_process_request_result* out_process_result);
TRACEABLE_RET traceable_delete_process_request_result_data(
    traceable_process_request_result result);

/*
 * traceable_decode_protobuf decodes a raw protobuf buffer into a null
 *   terminated JSON string. The caller must free "out_string". depth controls
 *   the number of nested groups that are decoded.
 */
TRACEABLE_RET traceable_decode_protobuf(const char* blob, int length, int depth,
                                        char** out_string);

TRACEABLE_RET traceable_is_content_type_capturable(
    const char* media_type, const char** supported_content_types,
    int supported_content_types_size, int* out_should_capture);

#ifdef __cplusplus
}
#endif

#endif