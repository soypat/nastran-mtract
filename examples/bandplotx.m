function  [] = bandplotx(elementos,nodos,variable, varargin)
% BANDPLOTX  FEA field variable grapher. 
% CC-BY-NC-SA Patricio Whittingslow 2019.
%
% Accepts elements:
% Hexahedral      8/20   nodes
% Tetrahedral     4/10   nodes
% Quadrilateral   4/8/9  nodes  (2D)
% Triangular      3/6    nodes  (2D)
% Typical element numbering, Q9 example given:
%  4---7---3
%  |       |
%  8   9   6
%  |       |
%  1---5---2
%
% BANDPLOTX(ELEMENTS, NODES, FIELDVAR) where
%
% ELEMENTS: Element connectivity matrix. Has as many rows as there are
%   elements.
% NODES: Node coordinate matrix. Columns are x, y, z values.
% FIELDVAR:  Must have same size as ELEMENTS. Values correspond to element 
%   nodes
%
% BANDPLOTX(ELEMENTS, NODES, FIELDVAR, NAME, VALUE) configures plotting
%   parameters with one or more name-value pair arguments:
%
% 'FaceOpacity'  - Opacity of faces of elements. 0 to 1 value
% 'NameNodes'    - Adds node numbering text to plot. Boolean
% 'NameElements' - Adds element numbers to plot. Boolean
% 'Coloring'     - Specifies colormap. ie. 'ADINA', jet. See help hsv
% 'LineColor'    - Sets rgb value element delimeter. Set to 'none' for no
%                  lines. 'Red' and others work
% 'FmtLegend'    - Format legend value. See help sprintf
% 'NColors'      - Number of colors. 'ADINA' overrides this
% 'NTicks'       - Number of values printed on color legend
% 'Optimize'     - Optimizes plot time or view performance Use 'view' 
%                  for many element plots
%
% Examples: 
%   BANDPLOTX(elem, nod, stress) % Simple plot no parameters
%
%   BANDPLOTX(elem, nod, stress, 'NameNodes', true, 'LineColor', 'None')
%   BANDPLOTX(elem, nod, temperature, 'Coloring', hot)

Nparam = nargin - 3;
if mod(Nparam,2)~=0 && Nparam > 0 %
    error('Pass parameters as groups of 2. Name-Value pair.')
end
%% Declare Default Parameters in Title Case / Pascal Case
FmtLegend = '%6.5E';
OptimizePlot = true;
NameNodes = false;
NameElem = false;
Coloring = jet;
NColors = 10; % numero de colores graficados
NTicks = NColors+1;
LineColor = [0 0 0];

FaceOpacity = 1; % transparencia de caras de elementos. 1 es opaco

%% Parameter setting
for arg = 1:2:Nparam
    switch varargin{arg}
        case {'FaceOpacity'}
            FaceOpacity = varargin{arg+1};
        case {'FmtLegend'}
            FmtLegend = varargin{arg+1};
        case {'Optimize'}
            if strcmpi(varargin{arg+1},'view')
                OptimizePlot = false;
            elseif ~strcmpi(varargin{arg+1},'plot')
                error('Bad value %s for name ''Optimize''. Use ''view'' or ''plot''.',varargin{arg+1})
            end
        case {'NameNodes'}
            if varargin{arg+1} % Early error checking
                NameNodes = varargin{arg+1};
            end
        case {'NameElements'}
            if varargin{arg+1} % Early error checking
                NameElem = varargin{arg+1};
            end
        case {'NColors'}
            NColors = varargin{arg+1};
        case {'NTicks'}
            NTicks = varargin{arg+1};
        case {'Coloring'}
            Coloring = varargin{arg+1};
            if ~strcmpi(Coloring,'adina') && isstring(Coloring)
                error('Coloring parameter must be an colormap or special string. Example: jet, hot etc.')
            end
        case {'LineColor'}
            LineColor = varargin{arg+1};
        otherwise
            error('Unknown name of parameter %s',varargin{arg})     
    end
end

%% Error check
[Nnod , Ndim]= size(nodos);
isValidVariable = all(size(variable) == size(elementos));
if ~isValidVariable
    error('size(ELEMENTS) must be equal to size(FIELDVAR)')
end
   

%% Limites de variable graficado
tol = 1E-5;
lims = [min(min(variable)) max(max(variable))];
if diff(lims) < lims(2)*tol
    lims = lims + [-1 1]*lims(2)*tol;
end


[Nelem, Nnodporelem] = size(elementos);
if Ndim==3
switch Nnodporelem
    case {10} %CTETRA
        vNod = [1 5 2 10 4 7;
                1 5 2 8 3 6
                1 6 3 9 4 7
                2 10 4 9 3 8];
    case {4} %T4 (LTETRA)
        vNod = [1 2 4
                1 2 3
                1 3 4
                2 4 3];
    case {8} %H8
        vNod = [1 2 3 4
                1 2 6 5
                2 6 7 3
                1 5 8 4
                5 6 7 8
                4 3 7 8];
    case {20} %CHEXA (H20)
       vNod = [1 9 2 10 3 11 4 12
                2 18 6 14 7 19 3 10
                4 12 1 17 5 16 8 20
                5 13 6 14 7 15 8 16
                1 9 2 18 6 13 5 17
                4 11 3 19 7 15 8 20];
    otherwise
        error('Elemento 3D desconocido.')
end
end

if Ndim == 2
    %FaceOpacity=1; % Override
    switch Nnodporelem
        case {3,4} % Caso lineales, CST y Q4
            vNod = 1:Nnodporelem;
        case {6} % LST
            vNod = [1 4 2 5 3 6];
        case {8,9} %Q8 y Q9
            vNod = [1 5 2 6 3 7 4 8];
        otherwise
            warning('Elemento 2D desconocido.')
            vNod = 1:Nnodporelem;
    end
end
[Nfaces, Nvertex] = size(vNod);
%% Graficador de superficies optimizada
thisAxes = gca;
faces = zeros(1,size(vNod,2));
skippedfaces=0;
for e = 1:Nelem
    elenod = elementos(e,:);
    for s = 1:Nfaces
        supindex = elementos(e,vNod(s,:));
        if OptimizePlot
            h = patch(thisAxes,'Faces',vNod(s,:),'Vertices',nodos(elenod,:),'FaceVertexCData',variable(e,:)'); 
            set(h,'FaceColor','interp','EdgeColor',LineColor,'CDataMapping','scaled');
            alpha(h,FaceOpacity);
        else
            if ~sum(sum( ismember(faces,sort(supindex),'rows'),2 ) == Nvertex)==1 % TODO: Optimize this line!
                faces = [faces;sort(supindex)];
                h = patch(thisAxes,'Faces',vNod(s,:),'Vertices',nodos(elenod,:),'FaceVertexCData',variable(e,:)'); 
                set(h,'FaceColor','interp','EdgeColor',LineColor,'CDataMapping','scaled');
                alpha(h,FaceOpacity);
            else
                skippedfaces=skippedfaces+1;
            end
        end
    end
    if NameElem
        xx = nodos(elenod,1);
        yy = nodos(elenod,2);
        if Ndim==2
            text(thisAxes,mean(xx),mean(yy),num2str(e),'VerticalAlignment','bottom','Color','b','FontSize',8);
        else
            zz = nodos(elenod,3);
            text(thisAxes,mean(xx),mean(yy),mean(zz),num2str(e),'VerticalAlignment','bottom','Color','b','FontSize',8);
        end
    end
end

%% Grafico nombre de nodos
if NameNodes
    for i = 1:Nnod
        xx(i) = nodos(i,1);
        yy(i) = nodos(i,2);
        if Ndim ==2
            text(thisAxes,xx(i),yy(i),num2str(i),'VerticalAlignment','bottom','Color','r','FontSize',8);
        else
            zz(i) = nodos(i,3);
            text(thisAxes,xx(i),yy(i),zz(i),num2str(i),'VerticalAlignment','bottom','Color','r','FontSize',8);
        end
     end
end
%% COLOR and OPACITY
adinaColor = [   42     0   255
                 0    42   255
                 0   128   255
                 0   212   255
                 0   255   212
                 0   255   128
                 0   255    42
                42   255     0
               128   255     0
               212   255     0
               255   212     0
               255   128     0
               255    42     0
               255     0    42]/255;

if strcmpi(Coloring, 'adina')
    colormap(thisAxes,adinaColor)
    NColors = size(adinaColor,1);NTicks = NColors+1;
else
    colormap(thisAxes,Coloring(floor(linspace(1,size(Coloring,1),NColors)),:));
end
caxis(thisAxes,lims);
%% Ticks
if NTicks > NColors
    NTicks = NColors;
end
ticks = lims(1):((diff(lims))/NTicks):lims(2);
tickLabels = cell(size(ticks));
for iTick = 1:length(ticks)
    tickLabels{iTick} = sprintf(FmtLegend, ticks(iTick));
end
%% Final Cleanup of axes
colorbar(thisAxes,'YTick',ticks,'YTickLabel',tickLabels);
if Ndim==3
    view(thisAxes,3)
end
if NameNodes
    xlabel(thisAxes,'x');ylabel(thisAxes,'y');zlabel(thisAxes,'z');
    maxyz=max(nodos);minxyz = min(nodos);
    xlim(thisAxes,[minxyz(1) maxyz(1)]);ylim([minxyz(2) maxyz(2)]);%zlim([minxyz(3) maxyz(3)])
else
    set(thisAxes,'XTick',[],'YTick',[],'ZTick',[],'XColor',[1 1 1],'YColor',[1 1 1],'ZColor',[1 1 1],'visible','on');
end
daspect(thisAxes,[1 1 1]);
box on
axis auto

end

