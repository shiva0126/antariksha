export interface ChartInput{date:string;time:string;lat:number;lon:number;tz:string}
export interface Graha{id:string;name:string;longitude:number;latitude:number;distance_au:number;rashi:string;rashi_degree:number;nakshatra:string;nakshatra_pada:number;retrograde:boolean;speed:number}
export interface ChartResponse{schema_version:number;input:ChartInput;ayanamsa:'lahiri';ascendant:{longitude:number;rashi:string;degree:number};grahas:Graha[];houses:{system:'whole_sign';cusps:number[]}}
