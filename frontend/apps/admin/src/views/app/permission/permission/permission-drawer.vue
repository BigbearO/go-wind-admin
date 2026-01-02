<script lang="ts" setup>
import type { ChangeEvent } from 'ant-design-vue/es/_util/EventInterface';

import { computed, reactive, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';
import { $t, $te } from '@vben/locales';

import lucide from '@iconify/json/json/lucide.json';
import { addCollection } from '@iconify/vue';
import { notification } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import {
  buildPermissionTree,
  PERMISSION_TYPE_NAME_MAP,
  statusList,
  usePermissionStore,
} from '#/stores';

const permissionStore = usePermissionStore();

addCollection(lucide);

const data = ref();

const titleSuffix = reactive({ title: '' });

const getTitle = computed(() =>
  data.value?.create
    ? $t('ui.modal.create', { moduleName: $t('page.permission.moduleName') })
    : $t('ui.modal.update', { moduleName: $t('page.permission.moduleName') }),
);

// const isCreate = computed(() => data.value?.create);

const [BaseForm, baseFormApi] = useVbenForm({
  showDefaultActions: false,
  // 所有表单项共用，可单独在表单内覆盖
  commonConfig: {
    formItemClass: 'col-span-2 md:col-span-1',
  },
  wrapperClass: 'grid-cols-2 gap-x-4',

  schema: [
    {
      component: 'RadioGroup',
      fieldName: 'type',
      label: $t('page.permission.type'),
      defaultValue: 'MENU',
      formItemClass: 'col-span-2 md:col-span-2',
      componentProps: {
        optionType: 'button',
        buttonStyle: 'solid',
        options: PERMISSION_TYPE_NAME_MAP,
      },
    },

    {
      component: 'Input',
      fieldName: 'name',
      label: $t('page.permission.name'),
      rules: 'required',
      componentProps() {
        return {
          placeholder: $t('ui.placeholder.input'),
          allowClear: true,
          addonAfter: titleSuffix.title,
          onChange({ target: { value } }: ChangeEvent) {
            titleSuffix.title = value && $te(value) ? $t(value) : '';
          },
        };
      },
    },
    {
      component: 'ApiTreeSelect',
      fieldName: 'parentId',
      label: $t('page.permission.parentId'),
      componentProps: {
        placeholder: $t('ui.placeholder.select'),
        class: 'w-full',
        showSearch: true,
        treeDefaultExpandAll: true,
        numberToString: true,
        allowClear: true,
        childrenField: 'children',
        labelField: 'name',
        valueField: 'id',
        treeNodeFilterProp: 'label',
        api: async () => {
          const fieldValue = baseFormApi.form.values;
          const result = await permissionStore.listPermission(undefined, {
            parentId: fieldValue.parentId,
            status: 'ON',
          });
          return result.items;
        },

        afterFetch: (data: any) => {
          return buildPermissionTree(data);
        },
      },
    },
    {
      component: 'InputNumber',
      fieldName: 'sortOrder',
      label: $t('page.permission.sortOrder'),
      componentProps: {
        placeholder: $t('ui.placeholder.input'),
        allowClear: true,
      },
    },
    {
      component: 'Input',
      fieldName: 'path',
      label: $t('page.permission.path'),
      rules: 'required',
      componentProps: {
        placeholder: $t('ui.placeholder.input'),
        allowClear: true,
      },
    },
    {
      component: 'RadioGroup',
      fieldName: 'status',
      defaultValue: 'ON',
      label: $t('ui.table.status'),
      rules: 'selectRequired',
      componentProps: {
        optionType: 'button',
        buttonStyle: 'solid',
        class: 'flex flex-wrap', // 如果选项过多，可以添加class来自动折叠
        options: statusList,
      },
    },
  ],
});

const [Drawer, drawerApi] = useVbenDrawer({
  onCancel() {
    drawerApi.close();
  },

  async onConfirm() {
    console.log('onConfirm');

    // 校验输入的数据
    const validate = await baseFormApi.validate();
    if (!validate.valid) {
      return;
    }

    setLoading(true);

    // 获取表单数据
    const values = await baseFormApi.getValues();

    console.log(getTitle.value, values);

    try {
      await (data.value?.create
        ? permissionStore.createPermission(values)
        : permissionStore.updatePermission(data.value.row.id, values));

      notification.success({
        message: data.value?.create
          ? $t('ui.notification.create_success')
          : $t('ui.notification.update_success'),
      });
    } catch {
      notification.error({
        message: data.value?.create
          ? $t('ui.notification.create_failed')
          : $t('ui.notification.update_failed'),
      });
    } finally {
      drawerApi.close();
      setLoading(false);
    }
  },

  onOpenChange(isOpen) {
    if (isOpen) {
      // 获取传入的数据
      data.value = drawerApi.getData<Record<string, any>>();

      // 为表单赋值
      baseFormApi.setValues(data.value?.row);

      setLoading(false);
    }
  },
});

function setLoading(loading: boolean) {
  drawerApi.setState({ loading });
}
</script>

<template>
  <Drawer :title="getTitle" class="w-full max-w-[800px]">
    <BaseForm class="mx-4" />
  </Drawer>
</template>
