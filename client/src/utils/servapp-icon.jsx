import { getFaviconURL } from "./routes";
import logogray from '../assets/images/icons/cosmos_gray.png';
import LazyLoad from 'react-lazyload';
import ImageWithPlaceholder from "../components/imageWithPlaceholder";

export const ServAppIcon = ({route, container, width, ...pprops}) => {
  return <LazyLoad width={width} height={width}>
    {(container && container.Labels["cosmos-icon"]) ? 
      <ImageWithPlaceholder src={container.Labels["cosmos-icon"]} {...pprops} width={width} height={width}></ImageWithPlaceholder> :(
        route ? <ImageWithPlaceholder src={getFaviconURL(route)} {...pprops} width={width} height={width}></ImageWithPlaceholder>
          : <ImageWithPlaceholder src={logogray} {...pprops} width={width} height={width}></ImageWithPlaceholder>)}
  </LazyLoad>;
};