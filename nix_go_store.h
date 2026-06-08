#ifndef NIX_GO_STORE_H
#define NIX_GO_STORE_H

#include <stddef.h>
#include <stdbool.h>

#include "nix_go_util.h"
#include "nix_api_store.h"
#include "nix_api_store/store_path.h"
#include "nix_api_store/derivation.h"

#ifdef __cplusplus
extern "C" {
#endif

typedef struct go_nix_store_param {
    const char *key;
    const char *value;
} go_nix_store_param;

typedef struct go_nix_store_params {
    const go_nix_store_param *items;
    size_t len;
} go_nix_store_params;

typedef struct go_nix_store_path_hash_part {
    unsigned char bytes[20];
} go_nix_store_path_hash_part;

typedef void (*go_nix_store_realise_callback)(
    void *userdata,
    const char *outname,
    const StorePath *out
);

typedef void (*go_nix_store_path_callback)(
    nix_c_context *context,
    void *userdata,
    const StorePath *store_path
);

nix_err go_nix_libstore_init(nix_c_context *ctx);
nix_err go_nix_libstore_init_no_load_config(nix_c_context *ctx);

Store *go_nix_store_open(
    nix_c_context *ctx,
    const char *uri,
    go_nix_store_params params
);

void go_nix_store_free(Store *store);

char *go_nix_store_get_uri(nix_c_context *ctx, Store *store);
char *go_nix_store_get_storedir(nix_c_context *ctx, Store *store);
char *go_nix_store_get_version(nix_c_context *ctx, Store *store);
char *go_nix_store_real_path(nix_c_context *ctx, Store *store, StorePath *path);

StorePath *go_nix_store_parse_path(nix_c_context *ctx, Store *store, const char *path);

StorePath *go_nix_store_path_clone(const StorePath *path);
void go_nix_store_path_free(StorePath *path);
char *go_nix_store_path_name(const StorePath *path);
nix_err go_nix_store_path_hash(
    nix_c_context *ctx,
    const StorePath *path,
    go_nix_store_path_hash_part *hash
);
StorePath *go_nix_store_create_from_parts(
    nix_c_context *ctx,
    const go_nix_store_path_hash_part *hash,
    const char *name,
    size_t name_len
);

bool go_nix_store_is_valid_path(
    nix_c_context *ctx,
    Store *store,
    const StorePath *path
);

nix_err go_nix_store_realise(
    nix_c_context *ctx,
    Store *store,
    StorePath *path,
    void *userdata,
    go_nix_store_realise_callback callback
);

nix_derivation *go_nix_derivation_from_json(
    nix_c_context *ctx,
    Store *store,
    const char *json
);

StorePath *go_nix_add_derivation(
    nix_c_context *ctx,
    Store *store,
    nix_derivation *derivation
);

nix_err go_nix_store_copy_closure(
    nix_c_context *ctx,
    Store *src_store,
    Store *dst_store,
    StorePath *path
);

nix_err go_nix_store_get_fs_closure(
    nix_c_context *ctx,
    Store *store,
    const StorePath *store_path,
    bool flip_direction,
    bool include_outputs,
    bool include_derivers,
    void *userdata,
    go_nix_store_path_callback callback
);

nix_derivation *go_nix_store_drv_from_store_path(
    nix_c_context *ctx,
    Store *store,
    const StorePath *path
);

StorePath *go_nix_store_query_path_from_hash_part(
    nix_c_context *ctx,
    Store *store,
    const char *hash
);

nix_err go_nix_store_copy_path(
    nix_c_context *ctx,
    Store *src_store,
    Store *dst_store,
    const StorePath *path,
    bool repair,
    bool check_sigs
);

#ifdef __cplusplus
}
#endif

#endif
