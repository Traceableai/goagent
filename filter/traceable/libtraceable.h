#ifndef LIBTRACEABLE_H
#define LIBTRACEABLE_H

#include <stdint.h>

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
  int enabled;
  const char* opa_server_url;
  const char* logging_dir;
  const char* logging_file_prefix;
  const char* cert_file;
  int use_secure_connection;
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
  int use_secure_connection;
  int64_t grpc_max_call_recv_msg_size;
} traceable_remote_config;

typedef struct {
  int enabled;
  traceable_opa_config opa_config;
  traceable_modsecurity_config modsecurity_config;
  traceable_rangeblocking_config rb_config;
  int evaluate_body;
  int skip_internal_request;
  int max_recursion_depth;
} traceable_blocking_config;

typedef struct {
  const char* service_name;
} traceable_agent_config;

typedef struct {
  int enabled;
} traceable_api_discovery_config;

typedef struct {
  int enabled;
} traceable_sampling_config;

typedef enum {
  TRACEABLE_NO_SPAN,
  TRACEABLE_BARE_SPAN,
  TRACEABLE_FULL_SPAN
} TRACEABLE_SPAN_TYPE;

typedef struct {
  int block;
  TRACEABLE_SPAN_TYPE span_type;
  int propagate;
  traceable_attributes attributes;
} traceable_process_request_result;

typedef struct {
  traceable_log_configuration log_config;
  traceable_remote_config remote_config;
  traceable_blocking_config blocking_config;
  traceable_agent_config agent_config;
  traceable_api_discovery_config api_discovery_config;
  traceable_sampling_config sampling_config;
} traceable_libtraceable_config;

typedef enum { TRACEABLE_SUCCESS, TRACEABLE_FAIL } TRACEABLE_RET;

typedef void* traceable_libtraceable;

typedef void* modsecurity_rule_engine;

typedef struct {
  char* rule_id;
  char* rule_message;
  char* match_message;
  char* match_attribute;
  int paranoia_level;
  char* rule_uuid;
} modsecurity_rule_match;

typedef struct {
  int count;
  modsecurity_rule_match* match_arr;
} modsecurity_rule_matches;

/*
 * Traceable api functions
 */
traceable_libtraceable_config init_libtraceable_config();
TRACEABLE_RET traceable_new_libtraceable(
    traceable_libtraceable_config libtraceable_config,
    traceable_libtraceable* out_libtraceable);
TRACEABLE_RET traceable_start_libtraceable(traceable_libtraceable libtraceable);
TRACEABLE_RET traceable_delete_libtraceable(
    traceable_libtraceable libtraceable);

/*
 * Process request for headers processing phase. Performs api naming,
 * sampling and blocking. All attributes will be evaluated. If
 * traceable_process_request_body() will be used, `http.request.body`
 * and `rpc.request.body` should not be set as input attributes. Those fields
 * will be evaluated twice. Use traceable_process_request() for single phase
 * processing.
 */
TRACEABLE_RET traceable_process_request_headers(
    traceable_libtraceable libtraceable, traceable_attributes attributes,
    traceable_process_request_result* out_process_result);
/*
 * Process request for body processing phase. Only performs blocking. All
 * attributes will be evaluated.
 */
TRACEABLE_RET traceable_process_request_body(
    traceable_libtraceable libtraceable, traceable_attributes attributes,
    traceable_process_request_result* out_process_result);
/*
 * Process request and perfom api naming, sampling and blocking in one phase.
 */
TRACEABLE_RET traceable_process_request(
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

TRACEABLE_RET traceable_barespan_attributes(traceable_libtraceable libtraceable,
                                            traceable_attributes attributes,
                                            traceable_attributes* out_result);

TRACEABLE_RET traceable_delete_barespan_attributes_result(
    traceable_attributes result);

TRACEABLE_RET modsecurity_new_rule_engine(
    const char* rules, modsecurity_rule_engine* out_rule_engine);

TRACEABLE_RET modsecurity_process_attributes(
    modsecurity_rule_engine rule_engine, traceable_attributes attributes,
    modsecurity_rule_matches* out_rule_matches);

TRACEABLE_RET modsecurity_cleanup_rule_matches(
    modsecurity_rule_matches rule_matches);

TRACEABLE_RET modsecurity_cleanup_rule_engine(
    modsecurity_rule_engine rule_engine);

#ifdef __cplusplus
}
#endif

#endif  // LIBTRACEABLE_H