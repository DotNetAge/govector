/*
 * GoVector C API - C 语言接口定义
 * 
 * 这是 GoVector 向量数据库的 C 语言绑定接口，
 * 允许从 Python、C++、Java 等语言调用 GoVector 核心功能。
 * 
 * @file govector_c.h
 * @package github.com/DotNetAge/govector/capi
 */

#ifndef GOVECTOR_C_H
#define GOVECTOR_C_H

#include <stddef.h>
#include <stdint.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

/* ============================================================================
 * 类型定义
 * ============================================================================ */

/**
 * @brief 距离度量类型
 */
typedef enum {
    DISTANCE_COSINE = 0,    ///< 余弦相似度 (推荐用于文本搜索)
    DISTANCE_EUCLID = 1,    ///< 欧几里得距离 (推荐用于图像搜索)
    DISTANCE_DOT = 2        ///< 点积 (推荐用于推荐系统)
} DistanceType;

/**
 * @brief HNSW 索引参数
 * 
 * HNSW (Hierarchical Navigable Small World) 是一种高效的近似最近邻搜索算法。
 */
typedef struct {
    int m;              ///< 每个节点的最大连接数 (默认：16)
    int ef_construction; ///< 构建时的搜索深度 (默认：200)
    int ef_search;      ///< 搜索时的深度 (默认：50)
    int k;              ///< 量化参数 (默认：2)
} HNSWParams;

/**
 * @brief 向量点数据结构
 */
typedef struct {
    const char* id;           ///< 点的唯一标识符
    uint64_t version;         ///< 版本号（用于并发控制）
    float* vector;            ///< 向量数据数组
    int vector_dim;           ///< 向量维度
    const char* payload_json; ///< Payload 元数据（JSON 格式）
} PointStruct;

/**
 * @brief 搜索结果结构
 */
typedef struct {
    const char* id;           ///< 点的唯一标识符
    uint64_t version;         ///< 版本号
    float score;              ///< 相似度分数（越接近 1 越相似）
    const char* payload_json; ///< Payload 元数据（JSON 格式）
} ScoredPoint;

/**
 * @brief 错误信息结构
 */
typedef struct {
    int code;                 ///< 错误码
    const char* message;      ///< 错误消息
} ErrorInfo;

/**
 * @brief 存储引擎句柄（不透明指针）
 */
typedef void* StorageHandle;

/**
 * @brief 集合句柄（不透明指针）
 */
typedef void* CollectionHandle;

/* ============================================================================
 * 错误码定义
 * ============================================================================ */

#define GOVECTOR_OK                      0   ///< 成功
#define GOVECTOR_ERROR_GENERAL           1   ///< 一般错误
#define GOVECTOR_ERROR_INVALID_PARAM     2   ///< 无效参数
#define GOVECTOR_ERROR_NOT_FOUND         3   ///< 资源未找到
#define GOVECTOR_ERROR_ALREADY_EXISTS    4   ///< 资源已存在
#define GOVECTOR_ERROR_DIM_MISMATCH      5   ///< 维度不匹配
#define GOVECTOR_ERROR_STORAGE_FAILURE   6   ///< 存储故障
#define GOVECTOR_ERROR_MEMORY_ALLOC      7   ///< 内存分配失败

/* ============================================================================
 * 存储引擎 API
 * ============================================================================ */

/**
 * @brief 创建新的存储引擎
 * 
 * @param db_path 数据库文件路径（例如："govector.db"）
 * @param error 输出错误信息（可为 NULL）
 * @return StorageHandle 存储句柄，失败返回 NULL
 * 
 * @example
 * ```c
 * ErrorInfo error;
 * StorageHandle storage = govector_storage_new("mydb.db", &error);
 * if (!storage) {
 *     fprintf(stderr, "Error: %s\n", error.message);
 * }
 * ```
 */
StorageHandle govector_storage_new(const char* db_path, ErrorInfo* error);

/**
 * @brief 关闭存储引擎
 * 
 * @param handle 存储句柄
 * @return int 错误码
 */
int govector_storage_close(StorageHandle handle);

/**
 * @brief 释放存储句柄
 * 
 * @param handle 存储句柄
 */
void govector_storage_free(StorageHandle handle);

/* ============================================================================
 * 集合管理 API
 * ============================================================================ */

/**
 * @brief 创建新的向量集合
 * 
 * @param name 集合名称
 * @param vector_dim 向量维度（必须为正整数）
 * @param metric 距离度量类型
 * @param use_hnsw 是否使用 HNSW 索引（true=快速近似搜索，false=精确搜索）
 * @param hnsw_params HNSW 参数配置
 * @param storage 存储句柄（可为 NULL 表示纯内存模式）
 * @param error 输出错误信息
 * @return CollectionHandle 集合句柄，失败返回 NULL
 * 
 * @example
 * ```c
 * HNSWParams params = {16, 200, 50, 2};
 * CollectionHandle col = govector_collection_create(
 *     "my_collection",
 *     128,
 *     DISTANCE_COSINE,
 *     true,
 *     params,
 *     storage,
 *     &error
 * );
 * ```
 */
CollectionHandle govector_collection_create(
    const char* name,
    int vector_dim,
    DistanceType metric,
    bool use_hnsw,
    HNSWParams hnsw_params,
    StorageHandle storage,
    ErrorInfo* error
);

/**
 * @brief 加载已存在的集合
 * 
 * @param name 集合名称
 * @param storage 存储句柄
 * @param error 输出错误信息
 * @return CollectionHandle 集合句柄，失败返回 NULL
 */
CollectionHandle govector_collection_load(
    const char* name,
    StorageHandle storage,
    ErrorInfo* error
);

/**
 * @brief 删除集合
 * 
 * @param name 集合名称
 * @param storage 存储句柄
 * @param error 输出错误信息
 * @return int 错误码
 */
int govector_collection_drop(
    const char* name,
    StorageHandle storage,
    ErrorInfo* error
);

/**
 * @brief 释放集合句柄
 * 
 * @param handle 集合句柄
 */
void govector_collection_free(CollectionHandle handle);

/* ============================================================================
 * 点操作 API
 * ============================================================================ */

/**
 * @brief 插入或更新向量点
 * 
 * @param handle 集合句柄
 * @param points 点数组
 * @param count 点的数量
 * @param error 输出错误信息
 * @return int 错误码
 * 
 * @example
 * ```c
 * PointStruct points[2];
 * // 初始化 points...
 * int ret = govector_collection_upsert(handle, points, 2, &error);
 * ```
 */
int govector_collection_upsert(
    CollectionHandle handle,
    const PointStruct* points,
    int count,
    ErrorInfo* error
);

/**
 * @brief 向量相似度搜索
 * 
 * @param handle 集合句柄
 * @param query_vector 查询向量
 * @param vector_dim 向量维度
 * @param top_k 返回结果数量
 * @param results 输出结果数组（需要调用 govector_scored_points_free 释放）
 * @param result_count 输出结果数量
 * @param error 输出错误信息
 * @return int 错误码
 * 
 * @example
 * ```c
 * ScoredPoint* results;
 * int count;
 * int ret = govector_collection_search(
 *     handle,
 *     query_vector,
 *     128,
 *     10,
 *     &results,
 *     &count,
 *     &error
 * );
 * 
 * // 处理结果...
 * 
 * // 释放内存
 * govector_scored_points_free(results, count);
 * ```
 */
int govector_collection_search(
    CollectionHandle handle,
    const float* query_vector,
    int vector_dim,
    int top_k,
    ScoredPoint** results,
    int* result_count,
    ErrorInfo* error
);

/**
 * @brief 删除向量点
 * 
 * @param handle 集合句柄
 * @param ids 点 ID 数组
 * @param count ID 数量
 * @param deleted_count 输出实际删除的数量
 * @param error 输出错误信息
 * @return int 错误码
 */
int govector_collection_delete(
    CollectionHandle handle,
    const char** ids,
    int count,
    int* deleted_count,
    ErrorInfo* error
);

/**
 * @brief 获取集合中点的数量
 * 
 * @param handle 集合句柄
 * @return int 点的数量
 */
int govector_collection_count(CollectionHandle handle);

/* ============================================================================
 * 内存管理 API
 * ============================================================================ */

/**
 * @brief 释放点数组
 * 
 * @param points 点数组
 * @param count 点的数量
 */
void govector_points_free(PointStruct* points, int count);

/**
 * @brief 释放搜索结果数组
 * 
 * @param results 结果数组
 * @param count 结果数量
 */
void govector_scored_points_free(ScoredPoint* results, int count);

/**
 * @brief 释放错误信息
 * 
 * @param error 错误信息结构
 */
void govector_error_free(ErrorInfo* error);

#ifdef __cplusplus
}
#endif

#endif /* GOVECTOR_C_H */
