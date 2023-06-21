import { getFaviconURL } from "./routes";
import logogray from '../assets/images/icons/cosmos_gray.png';

export const ServAppIcon = ({route, container, ...pprops}) => {
  return (container && container.Labels["cosmos-icon"]) ? 
    <img src={container.Labels["cosmos-icon"]} {...pprops}></img> :
    route ? <img src={getFaviconURL(route)} {...pprops}></img>
  : <img src={logogray} {...pprops}></img>
};