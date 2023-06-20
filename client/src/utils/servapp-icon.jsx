import { getFaviconURL } from "./routes";

export const ServAppIcon = ({route, container, ...pprops}) => {
  return (container && container.Labels["cosmos-icon"]) ? 
    <img src={container.Labels["cosmos-icon"]} {...pprops}></img>
  : <img src={getFaviconURL(route)} {...pprops}></img>
};