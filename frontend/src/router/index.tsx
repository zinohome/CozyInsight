import { createBrowserRouter } from 'react-router-dom';
import App from '../App';
import DatasourceList from '../pages/datasource';
import DatasetList from '../pages/dataset';
import ChartList from '../pages/chart';
import ChartEditor from '../pages/chart/ChartEditor';

import DashboardList from '../pages/dashboard';
import DashboardEditor from '../pages/dashboard/DashboardEditor';

const router = createBrowserRouter([
    {
        path: '/',
        element: <App />,
        children: [
            {
                path: '/',
                element: <DashboardList />,
            },
            {
                path: '/dashboard',
                element: <DashboardList />,
            },
            {
                path: '/dashboard/edit/:id',
                element: <DashboardEditor />,
            },
            {
                path: '/datasource',
                element: <DatasourceList />,
            },
            {
                path: '/dataset',
                element: <DatasetList />,
            },
            {
                path: '/chart',
                element: <ChartList />,
            },
            {
                path: '/chart/create',
                element: <ChartEditor />,
            },
            {
                path: '/chart/edit/:id',
                element: <ChartEditor />,
            },
        ],
    },
]);

export default router;
