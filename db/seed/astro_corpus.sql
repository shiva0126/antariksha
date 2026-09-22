INSERT INTO astro_corpus(doc_type,key,language,title,body,source) VALUES
 ('yoga','gajakesari','en','Gajakesari Yoga','Jupiter in a kendra from the Moon is traditionally associated with learning, judgment and supportive counsel. Interpret strength through dignity and affliction; it is a tendency, not a guarantee.','BPHS-inspired interpretation'),
 ('yoga','budha-aditya','en','Budha-Aditya Yoga','Sun and Mercury in one sign is traditionally associated with intellect, communication and analytical expression. Combustion changes emphasis but does not erase the geometric fact.','BPHS-inspired interpretation'),
 ('yoga','ruchaka','en','Ruchaka Mahapurusha Yoga','Mars in its own or exaltation sign in a kendra is traditionally associated with initiative, courage and disciplined action.','BPHS-inspired interpretation'),
 ('dignity','exalted','en','Exaltation','An exalted graha has a strong classical sign dignity. The expression still depends on house, aspects and the rest of the chart.','Classical Jyotisha summary')
ON CONFLICT DO NOTHING;
