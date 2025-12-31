import { computed } from 'vue';

import { $t } from '@vben/locales';

import { defineStore } from 'pinia';

import {
  createRoleServiceClient,
  type userservicev1_Role as Role,
  type userservicev1_Role_DataScope as Role_DataScope,
  type userservicev1_Role_Type as Role_Type,
} from '#/generated/api/admin/service/v1';
import { makeQueryString, makeUpdateMask } from '#/utils/query';
import { type Paging, requestClientRequestHandler } from '#/utils/request';

export const useRoleStore = defineStore('role', () => {
  const service = createRoleServiceClient(requestClientRequestHandler);

  /**
   * 查询角色列表
   */
  async function listRole(
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
   * 获取角色
   */
  async function getRole(id: number) {
    return await service.Get({ id });
  }

  /**
   * 创建角色
   */
  async function createRole(values: object) {
    return await service.Create({
      // @ts-ignore proto generated code is error.
      data: {
        ...values,
        children: [],
      },
    });
  }

  /**
   * 更新角色
   */
  async function updateRole(id: number, values: object) {
    if ('id' in values) delete values.id;

    return await service.Update({
      id,
      // @ts-ignore proto generated code is error.
      data: {
        ...values,
        children: [],
      },
      // @ts-ignore proto generated code is error.
      updateMask: makeUpdateMask(Object.keys(values ?? [])),
    });
  }

  /**
   * 删除角色
   */
  async function deleteRole(id: number) {
    return await service.Delete({ id });
  }

  function $reset() {}

  return {
    $reset,
    listRole,
    getRole,
    createRole,
    updateRole,
    deleteRole,
  };
});

/**
 * 在角色树中查找角色
 * @param list
 * @param id
 */
export const findRole = (list: Role[], id: number): null | Role | undefined => {
  for (const item of list) {
    // eslint-disable-next-line eqeqeq
    if (item.id == id) {
      return item;
    }

    if (item.children && item.children.length > 0) {
      const found = findRole(item.children, id);
      if (found) return found;
    }
  }

  return null;
};

// 角色类型-颜色映射常量（贴合系统/自定义角色的业务语义）
const ROLE_TYPE_COLOR_MAP = {
  SYSTEM: '#165DFF', // 系统角色：深蓝色（系统内置、权威、不可修改）
  CUSTOM: '#00B42A', // 自定义角色：企业绿（业务配置、灵活、可修改）
  DEFAULT: '#C9CDD4', // 未知角色类型：浅灰色（中性、无倾向）
} as const;

/**
 * 角色类型映射对应颜色
 * @param roleType 角色类型（SYSTEM/自定义角色）
 * @returns 标准化十六进制颜色值
 */
export function roleTypeToColor(roleType: Role_Type): string {
  // 类型安全断言 + 兜底默认色，避免无匹配值报错
  return (
    ROLE_TYPE_COLOR_MAP[roleType as keyof typeof ROLE_TYPE_COLOR_MAP] ||
    ROLE_TYPE_COLOR_MAP.DEFAULT
  );
}

// ---------------------- 数据范围颜色映射 ----------------------
// 数据范围-颜色映射常量（按权限范围从大到小匹配差异化色值）
const DATA_SCOPE_COLOR_MAP = {
  ALL: '#F53F3F', // 全部数据：红色（最大权限、高危、核心管控）
  UNIT_AND_CHILD: '#165DFF', // 本单位及子单位：深蓝色（大范围、层级化权限）
  UNIT_ONLY: '#FF7D00', // 仅本单位：橙色（中等范围、核心业务权限）
  SELECTED_UNITS: '#722ED1', // 指定单位：紫色（灵活范围、自定义权限）
  SELF: '#86909C', // 仅自己：中灰色（最小范围、基础权限）
  DEFAULT: '#C9CDD4', // 未知数据范围：浅灰色（中性、无倾向）
} as const;

/**
 * 数据范围映射对应颜色
 * @param dataScope 数据范围（ALL/SELF/UNIT_ONLY/UNIT_AND_CHILD/SELECTED_UNITS）
 * @returns 标准化十六进制颜色值
 */
export function dataScopeToColor(dataScope: Role_DataScope): string {
  return (
    DATA_SCOPE_COLOR_MAP[dataScope as keyof typeof DATA_SCOPE_COLOR_MAP] ||
    DATA_SCOPE_COLOR_MAP.DEFAULT
  );
}

export const roleTypeList = computed(() => [
  { label: $t('enum.role.type.SYSTEM'), value: 'SYSTEM' },
  { label: $t('enum.role.type.CUSTOM'), value: 'CUSTOM' },
]);

export const roleDataScopeList = computed(() => [
  { label: $t('enum.role.dataScope.ALL'), value: 'ALL' },
  { label: $t('enum.role.dataScope.UNIT_AND_CHILD'), value: 'UNIT_AND_CHILD' },
  { label: $t('enum.role.dataScope.UNIT_ONLY'), value: 'UNIT_ONLY' },
  { label: $t('enum.role.dataScope.SELECTED_UNITS'), value: 'SELECTED_UNITS' },
  { label: $t('enum.role.dataScope.SELF'), value: 'SELF' },
]);

/**
 * 角色类型转名称
 * @param roleType
 */
export function roleTypeToName(roleType: Role_Type) {
  const values = roleTypeList.value;
  const matchedItem = values.find((item) => item.value === roleType);
  return matchedItem ? matchedItem.label : '';
}

/**
 * 角色数据范围转名称
 * @param dataScope
 */
export function roleDataScopeToName(dataScope: Role_DataScope) {
  const values = roleDataScopeList.value;
  const matchedItem = values.find((item) => item.value === dataScope);
  return matchedItem ? matchedItem.label : '';
}
