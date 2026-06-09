#ifndef NIX_GO_MAIN_H
#define NIX_GO_MAIN_H

#include "nix_go_util.h"
#include "nix_api_main.h"

#ifdef __cplusplus
extern "C" {
#endif

nix_err go_nix_init_plugins(nix_c_context *ctx);
nix_err go_nix_set_log_format(nix_c_context *ctx, const char *format);

#ifdef __cplusplus
}
#endif

#endif
