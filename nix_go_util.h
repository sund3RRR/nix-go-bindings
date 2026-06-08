#ifndef NIX_GO_UTIL_H
#define NIX_GO_UTIL_H

#include "nix_api_util.h"

#ifdef __cplusplus
extern "C" {
#endif

nix_c_context *go_nix_c_context_create(void);
void go_nix_c_context_free(nix_c_context *ctx);

nix_err go_nix_libutil_init(nix_c_context *ctx);

char *go_nix_setting_get(nix_c_context *ctx, const char *key);
nix_err go_nix_setting_set(nix_c_context *ctx, const char *key, const char *value);

char *go_nix_version_get(void);

char *go_nix_err_msg(nix_c_context *ctx, const nix_c_context *read_ctx);
char *go_nix_err_info_msg(nix_c_context *ctx, const nix_c_context *read_ctx);
char *go_nix_err_name(nix_c_context *ctx, const nix_c_context *read_ctx);
nix_err go_nix_err_code(const nix_c_context *ctx);
nix_err go_nix_set_err_msg(nix_c_context *ctx, nix_err err, const char *msg);
void go_nix_clear_err(nix_c_context *ctx);

nix_err go_nix_set_verbosity(nix_c_context *ctx, nix_verbosity level);

void go_nix_string_free(char *s);

#ifdef __cplusplus
}
#endif

#endif
