import { defineStore } from 'pinia';

import {
  createPermissionServiceClient,
  type Permission,
  type Permission_Type,
} from '#/generated/api/admin/service/v1';
import { makeQueryString, makeUpdateMask } from '#/utils/query';
import { type Paging, requestClientRequestHandler } from '#/utils/request';

export const usePermissionStore = defineStore('permission', () => {
  const service = createPermissionServiceClient(requestClientRequestHandler);

  /**
   * 查询权限列表
   */
  async function listPermission(
    paging?: Paging,
    formValues?: null | object,
    fieldMask?: null | string,
    orderBy?: null | string[],
  ) {
    const noPaging =
      paging?.page === undefined && paging?.pageSize === undefined;
    return await service.List({
      // @ts-ignore proto generated code is error.
      fieldMask,
      orderBy: orderBy ?? [],
      query: makeQueryString(formValues ?? null),
      page: paging?.page,
      pageSize: paging?.pageSize,
      noPaging,
    });
  }

  /**
   * 获取权限
   */
  async function getPermission(id: number) {
    return await service.Get({ id });
  }

  /**
   * 创建权限
   */
  async function createPermission(values: object) {
    return await service.Create({
      // @ts-ignore proto generated code is error.
      data: {
        ...values,
      },
    });
  }

  /**
   * 更新权限
   */
  async function updatePermission(id: number, values: object) {
    return await service.Update({
      id,
      // @ts-ignore proto generated code is error.
      data: {
        ...values,
      },
      // @ts-ignore proto generated code is error.
      updateMask: makeUpdateMask(Object.keys(values ?? [])),
    });
  }

  /**
   * 删除权限
   */
  async function deletePermission(id: number) {
    return await service.Delete({ id });
  }

  function $reset() {}

  return {
    $reset,
    listPermission,
    getPermission,
    createPermission,
    updatePermission,
    deletePermission,
  };
});

interface PermissionTreeDataNode {
  key: number | string; // 节点唯一标识（父节点用module，子节点用api.id）
  title: string; // 节点显示文本（父节点用module，子节点用api.name）
  children?: PermissionTreeDataNode[]; // 子节点（仅父节点有）
  disabled?: boolean;
  permission?: Permission;
}

export function convertPermissionToTree(
  rawApiList: Permission[],
): PermissionTreeDataNode[] {
  const moduleMap = new Map<string, Permission[]>();
  rawApiList.forEach((api) => {
    const moduleName = api.name ?? '';
    if (!moduleMap.has(moduleName)) {
      moduleMap.set(moduleName, []);
    }
    moduleMap.get(moduleName)?.push(api);
  });

  return [...moduleMap.entries()].map(([moduleName, apiList]) => ({
    key: `module-${moduleName}`,
    title: moduleName,
    children: apiList.map((permission, index) => ({
      key: permission.id ?? `permission-default-${index}`,
      title: `${permission.name}（${permission.method}）`,
      permission,
    })),
  }));
}

// 权限类型-颜色映射常量
const PERMISSION_TYPE_COLOR_MAP = {
  MENU: '#165DFF', // 菜单权限：深蓝色（核心导航、层级化属性）
  BUTTON: '#00B42A', // 按钮权限：企业绿（操作级、功能点属性）
  API: '#722ED1', // 接口权限：深紫色（底层、技术级属性）
  DATA: '#FF7D00', // 数据权限：橙色（范围管控、业务核心属性）
  PAGE: '#14C9C9', // 页面权限：天蓝色（访问级、视图层属性）
  OTHER: '#86909C', // 其他权限：中灰色（中性、非核心属性）
  DEFAULT: '#C9CDD4', // 未知权限：浅灰色（无语义倾向）
} as const;

/**
 * 权限类型映射对应颜色
 * @param permissionType 权限类型（MENU/BUTTON/API/DATA/PAGE/OTHER）
 * @returns 标准化十六进制颜色值
 */
export function permissionTypeToColor(permissionType: Permission_Type): string {
  // 类型安全断言 + 兜底默认色，避免类型不匹配导致的异常
  return (
    PERMISSION_TYPE_COLOR_MAP[
      permissionType as keyof typeof PERMISSION_TYPE_COLOR_MAP
    ] || PERMISSION_TYPE_COLOR_MAP.DEFAULT
  );
}
