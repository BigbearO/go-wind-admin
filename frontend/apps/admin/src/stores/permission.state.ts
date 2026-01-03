import { computed } from 'vue';

import { $t } from '@vben/locales';

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

  async function syncApiResources() {
    return await service.SyncApiResources({});
  }

  async function syncMenus() {
    return await service.SyncMenus({});
  }

  function $reset() {}

  return {
    $reset,
    listPermission,
    getPermission,
    createPermission,
    updatePermission,
    deletePermission,
    syncApiResources,
    syncMenus,
  };
});

// 权限类型-颜色映射常量
const PERMISSION_TYPE_COLOR_MAP = {
  CATALOG: '#4A90E2', // 目录权限
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

// 权限类型-名称映射常量
export const PERMISSION_TYPE_NAME_MAP = computed(() => [
  { label: $t('enum.permission.type.CATALOG'), value: 'CATALOG' },
  { label: $t('enum.permission.type.MENU'), value: 'MENU' },
  { label: $t('enum.permission.type.PAGE'), value: 'PAGE' },
  { label: $t('enum.permission.type.BUTTON'), value: 'BUTTON' },
  { label: $t('enum.permission.type.API'), value: 'API' },
  { label: $t('enum.permission.type.DATA'), value: 'DATA' },
]);

export function permissionTypeToName(permissionType: Permission_Type): string {
  const values = PERMISSION_TYPE_NAME_MAP.value;
  const matchedItem = values.find((item) => item.value === permissionType);
  return matchedItem ? matchedItem.label : '';
}

/** 遍历菜单子节点
 * @param nodes 节点列表
 * @param parent 父节点
 * @return 是否找到并添加
 */
export function travelPermissionChild(
  nodes: Permission[] | undefined,
  parent: Permission,
): boolean {
  if (nodes === undefined) {
    return false;
  }

  if (parent.parentId === 0 || parent.parentId === undefined) {
    if (parent?.name) {
      parent.name = $t(parent?.name ?? '');
    }
    nodes.push(parent);
    return true;
  }

  for (const node of nodes) {
    if (node === undefined) {
      continue;
    }
    if (node.id === parent.parentId) {
      if (parent?.name) {
        parent.name = $t(parent?.name ?? '');
      }
      if (node.children !== undefined) {
        node.children.push(parent);
      }
      return true;
    }

    if (travelPermissionChild(node.children, parent)) {
      return true;
    }
  }

  return false;
}

/**
 * 构建菜单树
 * @param permissions 菜单列表
 * @return 菜单树
 */
export function buildPermissionTree(permissions: Permission[]): Permission[] {
  const tree: Permission[] = [];

  for (const menu of permissions) {
    if (!menu) {
      continue;
    }

    if (menu.parentId !== 0 && menu.parentId !== undefined) {
      continue;
    }

    if (menu?.name) {
      menu.name = $t(menu?.name ?? '');
    }
    tree.push(menu);
  }

  for (const menu of permissions) {
    if (!menu) {
      continue;
    }

    if (menu.parentId === 0 || menu.parentId === undefined) {
      continue;
    }

    if (travelPermissionChild(tree, menu)) {
      continue;
    }

    if (menu?.name) {
      menu.name = $t(menu?.name ?? '');
    }
    tree.push(menu);
  }

  return tree;
}
