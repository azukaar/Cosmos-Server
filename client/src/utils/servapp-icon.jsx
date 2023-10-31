import { getFaviconURL } from "./routes";
import logogray from '../assets/images/icons/cosmos_gray.png';
import LazyLoad from 'react-lazyload';

export const ServAppIcon = ({route, container, width, ...pprops}) => {
  return <LazyLoad width={width} height={width}>
    {(container && container.Labels["cosmos-icon"]) ? 
      <img src={container.Labels["cosmos-icon"]} {...pprops} width={width} height={width}></img> :(
        route ? <img src={getFaviconURL(route)} {...pprops} width={width} height={width}></img>
          : <img src={logogray} {...pprops} width={width} height={width}></img>)}
  </LazyLoad>;
};