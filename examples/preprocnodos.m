function [newelem, newnodos] = preprocnodos(elem, nodos)
% PREPROCNODOS elimina nodos que no se encuentran en la 
% matriz elem. 
%
% El tamaño de elem no es modificado, solo
% cambian sus numeros para quedar asociados a los índices
% de los nuevos elementos
    [Nn, Ndim] = size(nodos);
    [Ne, ~] = size(elem);
    Npresent = false(Nn,1);
    Npresent(elem(:)) = true;
    Nnn = sum(Npresent,1);
    Nnewnumber = nan(Nn,1);
    Ncount = 0;
    newnodos = nan(Nnn,Ndim);
    for n = 1:Nn
        if ~Npresent(n)
            Ncount = Ncount+1;
        else
            Nnewnumber(n) = n-Ncount;
            newnodos(Nnewnumber(n),:) = nodos(n,:);
        end
    end
    estretch = elem(:);
    for ne = 1:numel(estretch)
        estretch(ne) = Nnewnumber(estretch(ne));
    end
    newelem = reshape(estretch,Ne,[]);
end