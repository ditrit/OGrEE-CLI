USAGE:  camera [UnityATTRIBUTE]=[VALUE]   

Sends a command to the Unity Client to adjust the camera by 'UnityATTRIBUTE' according to 'VALUE'   

For more information please refer to:   
https://github.com/ditrit/OGrEE-3D/wiki/CLI-langage#Manipulate-camera

UNITY ATTRIBUTE DESCRIPTIONS

    wait - Define a delay between two camera translations in seconds
    move - Move the camera to the given point
    translate - Move the camera to the given destination. You can stack several destinations, the camera will move to each point in the given order


UNITY ATTRIBUTES AND VALUE FORMATS

    wait = integer
    move = vector3@vector2
    translate = vector3@vector2


EXAMPLE

    camera wait = 2
    camera move = [5,5,5]@[0,0]
    camera translate = [5,5,5]@[0,0]