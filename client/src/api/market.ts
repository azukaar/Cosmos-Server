import wrap, { type ApiResponse, type ApiFetch } from './wrap';

export interface MarketApp {
  Name: string;
  Description: string;
  Url: string;
  LongDescription: string;
  Tags: string[];
  Repository: string;
  Image: string;
  Screenshots: string[];
  Icon: string;
  Compose: string;
  SupportedArchitectures: string[];
  [key: string]: any;
}

export interface MarketResult {
  Showcase: MarketApp[];
  All: Record<string, MarketApp[]>;
}

export default function createMarketAPI(apiFetch: ApiFetch) {
  function list(): Promise<ApiResponse<MarketResult>> {
    return wrap(apiFetch('/cosmos/api/markets', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
    }))
  }

  return {
    list,
  };
}
