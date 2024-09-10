// project import
import NavCard from './NavCard';
import Navigation from './Navigation';
import SimpleBar from '../../../../components/third-party/SimpleBar';

// ==============================|| DRAWER CONTENT ||============================== //

const DrawerContent = () => (
    <SimpleBar
        sx={{
            '& .simplebar-content': {
                display: 'flex',
                flexDirection: 'column',
                // space between
                justifyContent: 'space-between',
                height: '100%',
            }
        }}
    >
        <Navigation />
    </SimpleBar>
);

export default DrawerContent;
