import PropTypes from 'prop-types';
import { useEffect, useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

// material-ui
import MuiBreadcrumbs from '@mui/material/Breadcrumbs';
import { Grid, Typography } from '@mui/material';

// project imports
import MainCard from '../MainCard';

// ==============================|| BREADCRUMBS ||============================== //

const Breadcrumbs = ({ navigation, title, ...others }) => {
    const { t } = useTranslation();
    const location = useLocation();
    const [main, setMain] = useState();
    const [item, setItem] = useState();
    let subItem = '';

    // extract /servapps/stack/:stack
    const subPath = location.pathname.split('/')[3];
    if(subPath && location.pathname.split('/')[4]) {
        subItem = <Typography variant="subtitle1" color="textPrimary">
            {location.pathname.split('/')[4]}
        </Typography>;
    }

    // set active item state
    const getCollapse = (menu) => {
        if (menu.children) {
            menu.children.filter((collapse) => {
                if (collapse.type && collapse.type === 'collapse') {
                    getCollapse(collapse);
                } else if (collapse.type && collapse.type === 'item') {
                    if (location.pathname.startsWith(collapse.url)) {
                        setMain(menu);
                        setItem(collapse);
                    }
                }
                return false;
            });
        }
    };

    useEffect(() => {
        navigation?.items?.map((menu) => {
            if (menu.type && menu.type === 'group') {
                getCollapse(menu);
            }
            return false;
        });
    });

    // only used for component demo breadcrumbs
    if (location.pathname === '/breadcrumbs') {
        location.pathname = '/dashboard/analytics';
    }

    let mainContent;
    let itemContent;
    let breadcrumbContent = <Typography />;
    let itemTitle = '';

    // collapse item
    if (main && main.type === 'collapse') {
        mainContent = (
            <Typography component={Link} to={document.location.pathname} variant="h6" sx={{ textDecoration: 'none' }} color="textSecondary">
                {t(main.title)}
            </Typography>
        );
    }

    // items
    if (item && item.type === 'item') {
        itemTitle = item.title;
        itemContent = (
            <Typography variant="subtitle1" color="textPrimary">
                {t(itemTitle)}
            </Typography>
        );

        // main
        if (item.breadcrumbs !== false) {
            breadcrumbContent = (
                <MainCard border={false} sx={{ mb: 3, bgcolor: 'transparent' }} {...others} content={false}>
                    <Grid container direction="column" justifyContent="flex-start" alignItems="flex-start" spacing={1}>
                        <Grid item>
                            <MuiBreadcrumbs aria-label="breadcrumb">
                                <Typography component={Link} to="/" color="textSecondary" variant="h6" sx={{ textDecoration: 'none' }}>
                                    Home
                                </Typography>
                                {mainContent}
                                {itemContent}
                                {subPath && <Typography variant="subtitle1" color="textPrimary">{subPath}</Typography>}
                                {subItem}
                            </MuiBreadcrumbs>
                        </Grid>
                        {title && (
                            <Grid item sx={{ mt: 2 }}>
                                <Typography variant="h5">{t(item.title)}</Typography>
                            </Grid>
                        )}
                    </Grid>
                </MainCard>
            );
        }
    }

    return breadcrumbContent;
};

Breadcrumbs.propTypes = {
    navigation: PropTypes.object,
    title: PropTypes.bool
};

export default Breadcrumbs;
