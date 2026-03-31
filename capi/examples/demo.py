#!/usr/bin/env python3
# govector/capi/examples/demo.py
# 完整的 GoVector Python 绑定演示程序

"""
GoVector Python Binding Demo
============================

这个演示程序展示了如何使用 GoVector 的 Python 绑定进行向量数据库操作。

功能包括:
- 创建和加载集合
- 插入向量数据
- 执行相似度搜索
- 删除数据
- 统计查询

运行方式:
    python3 demo.py

依赖:
    需要先运行 ./build.sh python 构建 Python 绑定
"""

import sys
import os
import json
from typing import List, Dict, Any

# 添加父目录到路径以便导入 govector 模块
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

try:
    import govector
except ImportError:
    print("❌ 错误：govector 模块未找到")
    print("\n请先构建 Python 绑定:")
    print("  cd govector/capi")
    print("  ./build.sh python")
    sys.exit(1)


def print_separator(char: str = "=", length: int = 70):
    """打印分隔线"""
    print(char * length)


def print_header(title: str):
    """打印标题"""
    print()
    print_separator("=", 70)
    print(f"  {title}")
    print_separator("=", 70)


def print_step(step_num: int, description: str):
    """打印步骤"""
    print(f"\n[步骤 {step_num}] {description}")
    print("-" * 70)


class VectorDatabaseClient:
    """
    GoVector 数据库客户端封装类
    
    提供高级 API 用于向量数据库操作
    """
    
    def __init__(self, db_path: str = "demo.db"):
        """
        初始化客户端
        
        Args:
            db_path: 数据库文件路径
        """
        self.db_path = db_path
        self.storage = None
        self.collections = {}
        self.error = govector.ErrorInfo()
        self.connected = False
    
    def connect(self) -> bool:
        """连接到数据库"""
        if self.connected:
            return True
        
        self.storage = govector.govector_storage_new(
            self.db_path,
            self.error
        )
        
        if not self.storage:
            print(f"❌ 连接失败：{self.error.message.decode('utf-8')}")
            govector.govector_error_free(self.error)
            return False
        
        self.connected = True
        print(f"✓ 成功连接到数据库：{self.db_path}")
        return True
    
    def disconnect(self):
        """断开连接并清理资源"""
        if not self.connected:
            return
        
        # 清理所有集合
        for handle in self.collections.values():
            govector.govector_collection_free(handle)
        self.collections.clear()
        
        # 关闭存储
        if self.storage:
            govector.govector_storage_close(self.storage)
            govector.govector_storage_free(self.storage)
            self.storage = None
        
        self.connected = False
        print("✓ 已断开连接")
    
    def create_collection(self, name: str, vector_dim: int, 
                         use_hnsw: bool = True,
                         m: int = 16,
                         ef_construction: int = 200,
                         ef_search: int = 50) -> bool:
        """
        创建集合
        
        Args:
            name: 集合名称
            vector_dim: 向量维度
            use_hnsw: 是否使用 HNSW 索引
            m: HNSW M 参数
            ef_construction: HNSW 构建参数
            ef_search: HNSW 搜索参数
        
        Returns:
            bool: 是否成功
        """
        if not self.connected:
            print("❌ 错误：未连接到数据库")
            return False
        
        hnsw_params = govector.HNSWParams()
        hnsw_params.m = m
        hnsw_params.ef_construction = ef_construction
        hnsw_params.ef_search = ef_search
        hnsw_params.k = 2
        
        handle = govector.govector_collection_create(
            name,
            vector_dim,
            govector.DISTANCE_COSINE,
            use_hnsw,
            hnsw_params,
            self.storage,
            self.error
        )
        
        if not handle:
            print(f"❌ 创建集合失败：{self.error.message.decode('utf-8')}")
            govector.govector_error_free(self.error)
            return False
        
        self.collections[name] = handle
        print(f"✓ 创建集合成功：{name} (维度={vector_dim}, HNSW={use_hnsw})")
        return True
    
    def load_collection(self, name: str) -> bool:
        """加载已存在的集合"""
        if not self.connected:
            return False
        
        handle = govector.govector_collection_load(
            name,
            self.storage,
            self.error
        )
        
        if not handle:
            print(f"❌ 加载集合失败：{self.error.message.decode('utf-8')}")
            govector.govector_error_free(self.error)
            return False
        
        self.collections[name] = handle
        print(f"✓ 加载集合成功：{name}")
        return True
    
    def insert_sample_data(self, collection_name: str) -> int:
        """
        插入示例数据
        
        由于 SWIG 包装的限制，这里简化了数据插入过程
        实际应用中需要创建 C struct 数组
        
        Args:
            collection_name: 集合名称
        
        Returns:
            int: 插入的数据量
        """
        if collection_name not in self.collections:
            print(f"❌ 集合不存在：{collection_name}")
            return 0
        
        # 注意：完整的 upsert 实现需要创建 C PointStruct 数组
        # 这里仅作为演示，实际项目中需要使用 ctypes 创建结构体
        
        print("ℹ️  提示：完整的 upsert 操作需要创建 C 结构体数组")
        print("   参考文档了解详细实现：SWIG_EXAMPLES_AND_BEST_PRACTICES.md")
        
        # 模拟插入成功
        return 3
    
    def search(self, collection_name: str, 
               query_vector: List[float], 
               top_k: int = 5) -> List[Dict[str, Any]]:
        """
        执行相似度搜索
        
        Args:
            collection_name: 集合名称
            query_vector: 查询向量
            top_k: 返回结果数量
        
        Returns:
            搜索结果列表
        """
        if collection_name not in self.collections:
            print(f"❌ 集合不存在：{collection_name}")
            return []
        
        handle = self.collections[collection_name]
        
        # 创建 C float 数组
        import ctypes
        vector_type = ctypes.c_float * len(query_vector)
        c_vector = vector_type(*query_vector)
        
        results_ptr = ctypes.POINTER(govector.ScoredPoint)()
        result_count = ctypes.c_int(0)
        
        ret = govector.govector_collection_search(
            handle,
            c_vector,
            len(query_vector),
            top_k,
            ctypes.byref(results_ptr),
            ctypes.byref(result_count),
            self.error
        )
        
        if ret != govector.GOVECTOR_OK:
            print(f"❌ 搜索失败：{self.error.message.decode('utf-8')}")
            govector.govector_error_free(self.error)
            return []
        
        # 转换结果为 Python 字典
        results = []
        if results_ptr and result_count.value > 0:
            for i in range(result_count.value):
                point = results_ptr[i]
                payload_json = point.payload_json.decode('utf-8') if point.payload_json else '{}'
                results.append({
                    'id': point.id.decode('utf-8'),
                    'score': point.score,
                    'payload': json.loads(payload_json)
                })
            
            # 释放内存
            govector.govector_scored_points_free(results_ptr, result_count)
        
        return results
    
    def delete(self, collection_name: str, ids: List[str]) -> int:
        """删除点"""
        if collection_name not in self.collections:
            return 0
        
        import ctypes
        handle = self.collections[collection_name]
        deleted_count = ctypes.c_int(0)
        
        # 创建 C 字符串数组
        c_ids = (ctypes.c_char_p * len(ids))(*[id.encode('utf-8') for id in ids])
        
        ret = govector.govector_collection_delete(
            handle,
            c_ids,
            len(ids),
            ctypes.byref(deleted_count),
            self.error
        )
        
        if ret != govector.GOVECTOR_OK:
            print(f"❌ 删除失败：{self.error.message.decode('utf-8')}")
            govector.govector_error_free(self.error)
            return 0
        
        return deleted_count.value
    
    def count(self, collection_name: str) -> int:
        """获取集合中点的数量"""
        if collection_name not in self.collections:
            return 0
        
        handle = self.collections[collection_name]
        return govector.govector_collection_count(handle)
    
    def list_collections(self) -> List[str]:
        """列出所有集合"""
        return list(self.collections.keys())


def demo_basic_operations():
    """演示基本操作"""
    print_header("GoVector Python 绑定基础演示")
    
    client = VectorDatabaseClient("demo.db")
    
    try:
        # 步骤 1: 连接
        print_step(1, "连接到数据库")
        if not client.connect():
            return
        print("✓ 连接成功")
        
        # 步骤 2: 创建集合
        print_step(2, "创建向量集合")
        success = client.create_collection(
            "demo_collection",
            vector_dim=3,
            use_hnsw=True,
            m=16,
            ef_construction=200,
            ef_search=50
        )
        if not success:
            return
        
        # 步骤 3: 获取统计信息
        print_step(3, "获取集合统计信息")
        count = client.count("demo_collection")
        print(f"✓ 集合包含 {count} 个点")
        
        # 步骤 4: 列出集合
        print_step(4, "列出所有集合")
        collections = client.list_collections()
        print(f"✓ 当前集合列表：{collections}")
        
        # 步骤 5: 搜索示例
        print_step(5, "执行相似度搜索（示例）")
        query = [0.1, 0.2, 0.3]
        print(f"查询向量：{query}")
        print("ℹ️  提示：需要先插入数据才能看到搜索结果")
        
        # 步骤 6: 清理
        print_step(6, "清理资源")
        
    finally:
        client.disconnect()
    
    print_header("演示完成")


def demo_advanced_features():
    """演示高级功能"""
    print_header("GoVector 高级功能演示")
    
    client = VectorDatabaseClient("advanced.db")
    
    try:
        # 连接
        if not client.connect():
            return
        
        # 创建多个集合
        print("\n创建多个集合...")
        client.create_collection("products", vector_dim=128)
        client.create_collection("users", vector_dim=64)
        client.create_collection("documents", vector_dim=256)
        
        # 列出集合
        collections = client.list_collections()
        print(f"\n已创建的集合：{collections}")
        
        # 获取统计
        for col_name in collections:
            count = client.count(col_name)
            print(f"  - {col_name}: {count} 个点")
        
        # 删除示例
        print("\n删除集合 'users'...")
        # 注意：实际的删除需要调用 drop_collection API
        
    finally:
        client.disconnect()
    
    print_header("高级功能演示完成")


def main():
    """主函数"""
    print("\n" + "=" * 70)
    print(" " * 20 + "GoVector Demo Program")
    print("=" * 70)
    print(f"\nGoVector 版本信息:")
    print(f"  - Python 绑定：0.1.0")
    print(f"  - 支持语言：Python, C++, Java, C#")
    print(f"  - 距离度量：Cosine, Euclidean, Dot Product")
    
    # 运行基础演示
    demo_basic_operations()
    
    # 运行高级演示
    demo_advanced_features()
    
    # 显示帮助信息
    print_header("更多信息")
    print("📚 完整文档:")
    print("   - SWIG_INTEGRATION_GUIDE.md")
    print("   - SWIG_IMPLEMENTATION.md")
    print("   - SWIG_EXAMPLES_AND_BEST_PRACTICES.md")
    print("   - SWIG_QUICK_REFERENCE.md")
    print("\n💡 提示:")
    print("   运行 './build.sh help' 查看所有可用命令")
    print("   访问 https://github.com/DotNetAge/govector 了解更多")
    
    print()
    print_separator("=", 70)
    print("演示结束，感谢使用！")
    print_separator("=", 70)
    print()


if __name__ == "__main__":
    main()
