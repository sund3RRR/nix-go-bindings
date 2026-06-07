#include "nix_go_store.h"

#include <stdlib.h>
#include <string.h>

typedef struct go_nix_string_capture {
    char *value;
} go_nix_string_capture;

static void go_nix_capture_string(const char *s, unsigned int n, void *userdata)
{
    go_nix_string_capture *capture = (go_nix_string_capture *)userdata;

    capture->value = (char *)malloc((size_t)n + 1);
    if (capture->value == NULL) {
        return;
    }

    memcpy(capture->value, s, (size_t)n);
    capture->value[n] = '\0';
}

static const char ***go_nix_params_pack(go_nix_store_params params)
{
    if (params.items == NULL || params.len == 0) {
        return NULL;
    }

    const char ***out = (const char ***)calloc(params.len + 1, sizeof(char **));
    if (out == NULL) {
        return NULL;
    }

    for (size_t i = 0; i < params.len; i++) {
        const char **pair = (const char **)calloc(3, sizeof(char *));
        if (pair == NULL) {
            for (size_t j = 0; j < i; j++) {
                free((void *)out[j]);
            }
            free(out);
            return NULL;
        }

        pair[0] = params.items[i].key;
        pair[1] = params.items[i].value;
        pair[2] = NULL;

        out[i] = pair;
    }

    out[params.len] = NULL;
    return out;
}

static void go_nix_params_free(const char ***params)
{
    if (params == NULL) {
        return;
    }

    for (size_t i = 0; params[i] != NULL; i++) {
        free((void *)params[i]);
    }

    free((void *)params);
}

nix_err go_nix_libstore_init(nix_c_context *ctx)
{
    return nix_libstore_init(ctx);
}

nix_err go_nix_libstore_init_no_load_config(nix_c_context *ctx)
{
    return nix_libstore_init_no_load_config(ctx);
}

Store *go_nix_store_open(
    nix_c_context *ctx,
    const char *uri,
    go_nix_store_params params
)
{
    const char ***packed = go_nix_params_pack(params);
    Store *store = nix_store_open(ctx, uri, packed);
    go_nix_params_free(packed);
    return store;
}

void go_nix_store_free(Store *store)
{
    nix_store_free(store);
}

static char *go_nix_get_string(
    nix_c_context *ctx,
    nix_err (*fn)(nix_c_context *, Store *, nix_get_string_callback, void *),
    Store *store
)
{
    go_nix_string_capture capture = {0};

    nix_err err = fn(ctx, store, go_nix_capture_string, &capture);
    if (err != NIX_OK) {
        free(capture.value);
        return NULL;
    }

    return capture.value;
}

char *go_nix_store_get_uri(nix_c_context *ctx, Store *store)
{
    return go_nix_get_string(ctx, nix_store_get_uri, store);
}

char *go_nix_store_get_storedir(nix_c_context *ctx, Store *store)
{
    return go_nix_get_string(ctx, nix_store_get_storedir, store);
}

char *go_nix_store_get_version(nix_c_context *ctx, Store *store)
{
    return go_nix_get_string(ctx, nix_store_get_version, store);
}

char *go_nix_store_real_path(nix_c_context *ctx, Store *store, StorePath *path)
{
    go_nix_string_capture capture = {0};

    nix_err err = nix_store_real_path(ctx, store, path, go_nix_capture_string, &capture);
    if (err != NIX_OK) {
        free(capture.value);
        return NULL;
    }

    return capture.value;
}

void go_nix_string_free(char *s)
{
    free(s);
}

StorePath *go_nix_store_parse_path(nix_c_context *ctx, Store *store, const char *path)
{
    return nix_store_parse_path(ctx, store, path);
}

bool go_nix_store_is_valid_path(nix_c_context *ctx, Store *store, const StorePath *path)
{
    return nix_store_is_valid_path(ctx, store, path);
}

nix_err go_nix_store_realise(
    nix_c_context *ctx,
    Store *store,
    StorePath *path,
    void *userdata,
    go_nix_store_realise_callback callback
)
{
    return nix_store_realise(ctx, store, path, userdata, callback);
}

nix_derivation *go_nix_derivation_from_json(
    nix_c_context *ctx,
    Store *store,
    const char *json
)
{
    return nix_derivation_from_json(ctx, store, json);
}

StorePath *go_nix_add_derivation(
    nix_c_context *ctx,
    Store *store,
    nix_derivation *derivation
)
{
    return nix_add_derivation(ctx, store, derivation);
}

nix_err go_nix_store_copy_closure(
    nix_c_context *ctx,
    Store *src_store,
    Store *dst_store,
    StorePath *path
)
{
    return nix_store_copy_closure(ctx, src_store, dst_store, path);
}

nix_err go_nix_store_get_fs_closure(
    nix_c_context *ctx,
    Store *store,
    const StorePath *store_path,
    bool flip_direction,
    bool include_outputs,
    bool include_derivers,
    void *userdata,
    go_nix_store_path_callback callback
)
{
    return nix_store_get_fs_closure(
        ctx,
        store,
        store_path,
        flip_direction,
        include_outputs,
        include_derivers,
        userdata,
        callback
    );
}

nix_derivation *go_nix_store_drv_from_store_path(
    nix_c_context *ctx,
    Store *store,
    const StorePath *path
)
{
    return nix_store_drv_from_store_path(ctx, store, path);
}

StorePath *go_nix_store_query_path_from_hash_part(
    nix_c_context *ctx,
    Store *store,
    const char *hash
)
{
    return nix_store_query_path_from_hash_part(ctx, store, hash);
}

nix_err go_nix_store_copy_path(
    nix_c_context *ctx,
    Store *src_store,
    Store *dst_store,
    const StorePath *path,
    bool repair,
    bool check_sigs
)
{
    return nix_store_copy_path(ctx, src_store, dst_store, path, repair, check_sigs);
}