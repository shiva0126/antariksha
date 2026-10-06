export const rashis=[['Mesha','मे','♈'],['Vrishabha','वृ','♉'],['Mithuna','मि','♊'],['Karka','क','♋'],['Simha','सिं','♌'],['Kanya','क','♍'],['Tula','तु','♎'],['Vrishchika','वृ','♏'],['Dhanu','ध','♐'],['Makara','म','♑'],['Kumbha','कुं','♒'],['Meena','मी','♓']]as const;
export const grahaGlyph:Record<string,string>={sun:'☉',moon:'☽',mars:'♂',mercury:'☿',jupiter:'♃',venus:'♀',saturn:'♄',rahu:'☊',ketu:'☋'};
/** Chart-mark colours for grahas (bars, dots), validated with the dataviz
 *  palette checker on the dark surface #0d1322: lightness band, chroma floor,
 *  CVD and normal-vision separation in dasha order, 3:1 contrast. Text never
 *  uses these; a coloured mark sits beside a name in text colour. */
export const grahaMark:Record<string,string>={ketu:'#8b6cf0',venus:'#e0569b',sun:'#d9731f',moon:'#3796d6',mars:'#d6453d',rahu:'#1fa38a',jupiter:'#b8860b',saturn:'#5a63d6',mercury:'#4fa34a'};
/** Glyph tints for planet symbols beside names (not for marks or text). */
export const grahaColor:Record<string,string>={sun:'#ffbd5a',moon:'#dfeaff',mars:'#ff635c',mercury:'#63d6ad',jupiter:'#f5cb73',venus:'#f5a9d0',saturn:'#9eaad8',rahu:'#76e6ec',ketu:'#bd8dff'};
