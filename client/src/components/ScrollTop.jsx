import PropTypes from 'prop-types';
import { useEffect, useRef } from 'react';
import { useLocation } from 'react-router-dom';

// ==============================|| NAVIGATION - SCROLL TO TOP ||============================== //

const isMarketRoute = (p) => {
    const segments = p.split('/').filter(Boolean);
    return segments.length >= 2 && segments[0] === 'cosmos-ui' && segments[1] === 'market-listing';
};

// The market listing is exactly /cosmos-ui/market-listing (no app segments)
const isMarketListing = (p) => {
    const segments = p.split('/').filter(Boolean);
    return segments.length === 2 && segments[0] === 'cosmos-ui' && segments[1] === 'market-listing';
};

const ScrollTop = ({ children }) => {
    const location = useLocation();
    const { pathname } = location;
    const prevPathname = useRef(pathname);
    const savedMarketScroll = useRef(0);

    // While on the market listing, track the user's scroll position so we can
    // restore it when the detail overlay is opened (the route change to
    // /market-listing/:store/:name resets the window scroll to 0, which would
    // otherwise jump the listing back to the top behind the overlay).
    useEffect(() => {
        if (isMarketListing(pathname)) {
            const onScroll = () => {
                savedMarketScroll.current = window.scrollY;
            };
            window.addEventListener('scroll', onScroll, { passive: true });
            return () => window.removeEventListener('scroll', onScroll);
        }
    }, [pathname]);

    useEffect(() => {
        const prevIsMarket = isMarketRoute(prevPathname.current);
        const nextIsMarket = isMarketRoute(pathname);
        const prevIsListing = isMarketListing(prevPathname.current);
        const nextIsListing = isMarketListing(pathname);

        const togglingMarketOverlay = prevIsMarket && nextIsMarket;

        prevPathname.current = pathname;

        if (togglingMarketOverlay) {
            // Opening (listing -> detail) or closing (detail -> listing): the
            // listing stays mounted behind the overlay. The route change makes
            // the browser reset the window scroll to 0 asynchronously (after
            // paint), so a single restore gets overwritten. Keep trying until
            // the scroll sticks, with a final fallback on popstate.
            const saved = savedMarketScroll.current;
            let attempts = 0;
            const tryRestore = () => {
                attempts++;
                window.scrollTo(0, saved);
                if (attempts < 10 && window.scrollY !== saved) {
                    requestAnimationFrame(tryRestore);
                }
            };
            requestAnimationFrame(tryRestore);
            window.addEventListener('popstate', tryRestore, { once: true });
            return;
        }

        window.scrollTo({
            top: 0,
            left: 0,
            behavior: 'smooth'
        });
    }, [pathname]);

    return children || null;
};

ScrollTop.propTypes = {
    children: PropTypes.node
};

export default ScrollTop;
