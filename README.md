# TP0: Docker + Comunicaciones + Concurrencia

Este repositorio implementa las resoluciones propuestas a los ejercicios del TP0 de la materia Sistemas Distribuidos 1 de la Facultad de Ingenieria, Universidad de Buenos Aires.  
  
A continuación se detallan las soluciones propuestas para cada ejercicio.

# Ejercicio 1
Para este ejercicio se pidio implementar un script que reciba por parametro la cantidad de clientes a generar, y el archivo `.yaml` a modificar para agregar los clientes.  
  
El script se puede correr con el siguiente comando:
```
./generar-compose.sh <archivo.yaml> <cantidad>
```
   
El algoritmo para generar el archivo esta implementado en Python y fue hecho mediante la modificacion y concatenacion de cadenas correspondientes al formato de un archivo `.yaml`.

# Ejercicio 2
En este ejercicio se tuvo que asociar los archivos de configuracion tanto para el cliente (`config.yaml`) como para el servidor (`config.ini`) para que estos no requieran levantar un nuevo container cada vez que se modificasen. Esto fue hecho con la creacion de un volumen y haciendo *bind mount* de los archivos de configuracion en el container.  
  