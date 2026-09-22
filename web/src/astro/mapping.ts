export const rad=(d:number)=>d*Math.PI/180;
export function wheelPoint(longitude:number,radius:number,cx=0,cy=0){const a=rad(-90-longitude);return{x:cx+radius*Math.cos(a),y:cy+radius*Math.sin(a)}}
export function domeDirection(longitude:number,latitude:number,radius=36):[number,number,number]{const l=rad(longitude),b=rad(latitude),c=Math.cos(b);return[c*Math.cos(l)*radius,Math.sin(b)*radius,c*Math.sin(l)*radius]}
