import type { RouteRecordRaw } from 'vue-router';

import { BasicLayout } from '#/layouts';
import { $t } from '#/locales';

const log: RouteRecordRaw[] = [
  {
    path: '/log',
    name: 'LogAuditManagement',
    component: BasicLayout,
    redirect: '/log/login',
    meta: {
      order: 2004,
      icon: 'lucide:activity',
      title: $t('menu.log.moduleName'),
      keepAlive: true,
      authority: ['platform_admin'],
    },
    children: [
      {
        path: 'login-audit-logs',
        name: 'LoginAuditLog',
        meta: {
          icon: 'lucide:user-lock',
          title: $t('menu.log.loginAuditLog'),
          authority: ['platform_admin'],
        },
        component: () => import('#/views/app/log/login_audit_log/index.vue'),
      },

      {
        path: 'operation-audit-logs',
        name: 'OperationAuditLog',
        meta: {
          icon: 'lucide:file-clock',
          title: $t('menu.log.operationAuditLog'),
          authority: ['platform_admin'],
        },
        component: () =>
          import('#/views/app/log/operation_audit_log/index.vue'),
      },
    ],
  },
];

export default log;
