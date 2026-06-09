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
    size_t key_len;
    const char *value;
    size_t value_len;
} go_nix_store_param;

typedef struct go_nix_store_params {
    const go_nix_store_param *items;
    size_t len;
} go_nix_store_params;

typedef struct go_nix_store_path_hash_part {
    unsigned char bytes[20];
} go_nix_store_path_hash_part;

typedef struct go_nix_store_realise_results go_nix_store_realise_results;
typedef struct go_nix_store_path_array go_nix_store_path_array;

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

go_nix_store_realise_results *go_nix_store_realise_to_array(
    nix_c_context *ctx,
    Store *store,
    StorePath *path
);
size_t go_nix_store_realise_results_count(const go_nix_store_realise_results *results);
char *go_nix_store_realise_results_outname(
    const go_nix_store_realise_results *results,
    size_t index
);
StorePath *go_nix_store_realise_results_path_clone(
    const go_nix_store_realise_results *results,
    size_t index
);
void go_nix_store_realise_results_free(go_nix_store_realise_results *results);

nix_derivation *go_nix_derivation_from_json(
    nix_c_context *ctx,
    Store *store,
    const char *json
);

nix_derivation *go_nix_derivation_clone(const nix_derivation *derivation);
void go_nix_derivation_free(nix_derivation *derivation);
char *go_nix_derivation_to_json(nix_c_context *ctx, const nix_derivation *derivation);

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

go_nix_store_path_array *go_nix_store_get_fs_closure_array(
    nix_c_context *ctx,
    Store *store,
    const StorePath *store_path,
    bool flip_direction,
    bool include_outputs,
    bool include_derivers
);
size_t go_nix_store_path_array_count(const go_nix_store_path_array *paths);
StorePath *go_nix_store_path_array_path_clone(
    const go_nix_store_path_array *paths,
    size_t index
);
void go_nix_store_path_array_free(go_nix_store_path_array *paths);

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
