import PropTypes from 'prop-types';

// material-ui
import { useTheme } from '@mui/material/styles';
import { Grid, Stack, Typography } from '@mui/material';

// project import
import MainCard from '../../components/MainCard';
import ComponentSkeleton from './ComponentSkeleton';

// ===============================|| SHADOW BOX ||=============================== //

function ShadowBox({ shadow }) {
    return (
        <MainCard border={false} sx={{ boxShadow: shadow }}>
            <Stack spacing={1} justifyContent="center" alignItems="center">
                <Typography variant="h6">boxShadow</Typography>
                <Typography variant="subtitle1">{shadow}</Typography>
            </Stack>
        </MainCard>
    );
}

ShadowBox.propTypes = {
    shadow: PropTypes.string.isRequired
};

// ===============================|| CUSTOM - SHADOW BOX ||=============================== //

function CustomShadowBox({ shadow, label, color, bgcolor }) {
    return (
        <MainCard border={false} sx={{ bgcolor: bgcolor || 'inherit', boxShadow: shadow }}>
            <Stack spacing={1} justifyContent="center" alignItems="center">
                <Typography variant="subtitle1" color={color}>
                    {label}
                </Typography>
            </Stack>
        </MainCard>
    );
}

CustomShadowBox.propTypes = {
    shadow: PropTypes.string.isRequired,
    color: PropTypes.string.isRequired,
    label: PropTypes.string.isRequired,
    bgcolor: PropTypes.string.isRequired
};

// ============================|| COMPONENT - SHADOW ||============================ //

const ComponentShadow = () => {
    const theme = useTheme();

    return (
        <ComponentSkeleton>
            <Grid container spacing={3}>
                <Grid item xs={12}>
                    <MainCard title="Basic Shadow" codeHighlight>
                        <Grid container spacing={3}>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="0" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="1" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="2" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="3" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="4" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="5" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="6" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="7" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="8" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="9" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="10" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="11" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="12" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="13" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="14" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="15" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="16" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="17" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="18" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="19" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="20" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="21" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="22" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="23" />
                            </Grid>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <ShadowBox shadow="24" />
                            </Grid>
                        </Grid>
                    </MainCard>
                </Grid>
                <Grid item xs={12}>
                    <MainCard title="Custom Shadow" codeHighlight>
                        <Grid container spacing={3}>
                            <Grid item xs={6} sm={4} md={3} lg={2}>
                                <CustomShadowBox shadow={theme.customShadows.z1} label="z1" color="inherit" />
                            </Grid>
                        </Grid>
                    </MainCard>
                </Grid>
            </Grid>
        </ComponentSkeleton>
    );
};

export default ComponentShadow;
