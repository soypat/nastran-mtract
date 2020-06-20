elem = load('CTETRA4-2.csv');
% Si tengo mas de una malla las junto:
% elem = [elem1;elem2;elem3];  % Junto todas las mallas de elementos
elem = elem(:,2:end);         % Excluyo la primera fila, es de numeracion
aux = load('nodos.csv');
% Nodos sera inicializada como una matriz no-numerica para no crear
% nodos sueltos en el origen de coordenadas
nodos = nan(max(aux(:,1)),3);
% Asigno los nodos a su lugar correspondiente en la matriz nodos.
% nodo 776 va ir a la fila 776, nodo 8 a la fila 8, etc.
nodos(aux(:,1),:) = aux(:,2:4);
% En este caso no hace falta pre-proceso, igual se muestra como se 
% usaría PREPROCNODOS para eliminar nodos 'sueltos'
[elem, nodos] = preprocnodos(elem,nodos);
% Puedo graficar para ver si se procesaron bien
bandplotx(elem, nodos, ones(size(elem))); 