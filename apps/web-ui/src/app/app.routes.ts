import { Route } from '@angular/router';
import { DashboardPage } from '@pages/dashboard-page';

export const appRoutes: Route[] = [
  {
    path: '',
    component: DashboardPage,
  },
  {
    path: '**',
    redirectTo: '',
  },
];
