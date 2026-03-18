const dashboardRoutes = [
  {
    path: '/dashboard',
    name: 'dashboard',
    icon: 'DashboardOutlined',
    routes: [
      {
        path: '/dashboard',
        redirect: '/dashboard/analysis',
      },
      {
        name: 'analysis',
        icon: 'BarChartOutlined',
        path: '/dashboard/analysis',
        component: './dashboard/analysis',
      },
      {
        name: 'monitor',
        icon: 'MonitorOutlined',
        path: '/dashboard/monitor',
        component: './dashboard/monitor',
      },
      {
        name: 'workplace',
        icon: 'HomeOutlined',
        path: '/dashboard/workplace',
        component: './dashboard/workplace',
      },
    ],
  },
];

export default dashboardRoutes;
