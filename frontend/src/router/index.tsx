import { createBrowserRouter } from 'react-router-dom';
import App from '../App';
import DatasourceList from '../pages/datasource';
import DatasourceCreate from '../pages/datasource/DatasourceCreate';
import DatasetList from '../pages/dataset';
import DatasetCreate from '../pages/dataset/DatasetCreate';
import ChartList from '../pages/chart';
import ChartEditor from '../pages/ChartEditor';
import DashboardList from '../pages/dashboard';
import DashboardEditor from '../pages/DashboardEditor';

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
                path: '/dashboard/create',
                element: <DashboardEditor />,
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
                path: '/datasource/create',
                element: <DatasourceCreate />,
            },
            {
                path: '/dataset',
                element: <DatasetList />,
            },
            {
                path: '/dataset/create',
                element: <DatasetCreate />,
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
