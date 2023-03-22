import pandas as pd
import json
import requests
import os
import sys


# Get a given data from a dictionary with position provided as a list
def get_from_dict(data_dict, map_list):    
    try:
        for k in map_list: 
            index_position = list(data_dict.keys()).index(k)
            data_dict = data_dict[k]
    except Exception as e:
        # print(e)
        return None, None
        
    return data_dict, index_position

# Check a given field in a JSON object for various mismatches, and recursively go down
# the object if needed
def field_checker(results, 
                 stats_dict, 
                 desired_field_key,
                 path_to_field,
                 desired_output, 
                 returned_output,
                 error_found):
    
    error_found = False
    
    desired_field, desired_index_position = get_from_dict(desired_output, path_to_field)
    
    if desired_field is None:
        print("Field {} not found in desired transaction".format(desired_field_key))
        sys.exit(1)
    
    output_field, output_index_position = get_from_dict(returned_output, path_to_field)                
    
    if output_field is None:
        results = pd.concat([results, pd.DataFrame({
            'problematicField': desired_field_key,
            'pathToField': ".".join(str(x) for x in path_to_field),
            'expectedValue': desired_field,
            'returnedValue': "Field not present in output"
        }, index=[0])], ignore_index=True)
        
        error_found = True
        stats_dict['fieldNotPresentErrors'] += 1
    
    elif type(desired_field) != type(output_field):
        results = pd.concat([results, pd.DataFrame({
            'problematicField': desired_field_key,
            'pathToField': ".".join(str(x) for x in path_to_field),
            'expectedValue': desired_field,
            'returnedValue': """Field in output has the wrong type, 
                                expected {} and got {}""".format(
                                                        type(desired_field), 
                                                        type(output_field))
        }, index=[0])], ignore_index=True)
        
        error_found = True
        stats_dict['wrongFieldTypeErrors'] += 1
            
    else:
        if desired_index_position != output_index_position:
            results = pd.concat([results, pd.DataFrame({
                    'problematicField': desired_field_key,
                    'pathToField': ".".join(str(x) for x in path_to_field),
                    'expectedValue': "Field should be in position " + str(desired_index_position) + " but is in position " + str(output_index_position),
                    'returnedValue': None
                }, index=[0])], ignore_index=True)
                
            error_found = True
            stats_dict['fieldInWrongPositionErrors'] += 1
                    
        # If field is a nested object, recursive call
        if type(desired_field) == dict:
            path_to_field.append(desired_field_key)
            
            print("Recursive calling field_checker with path_to_field = {}".format(path_to_field.join(".")))
            field_checker(results=results,
                          stats_dict=stats_dict, 
                          desired_field_key=desired_field_key,
                          path_to_field=path_to_field,
                          desired_output=desired_output, 
                          returned_output=returned_output, 
                          error_found=error_found)
        else:
            if desired_field != output_field:
                results = pd.concat([results, pd.DataFrame({
                    'problematicField': desired_field_key,
                    'pathToField': ".".join(str(x) for x in path_to_field),
                    'expectedValue': desired_field,
                    'returnedValue': output_field
                }, index=[0])], ignore_index=True)
                
                error_found = True
                stats_dict['valueMismatchErrors'] += 1
                
    return results, stats_dict, error_found


# Read JSON objects

with open("erigon_output.json", "r") as f:
    desired_output = json.load(f)['result']
    
with open("erigon_output1.json", "r") as f:
    returned_output = json.load(f)['result']


results = pd.DataFrame(columns=['problematicField',
                                'pathToField',
                                'expectedValue',
                                'returnedValue'])

stats = {"fieldNotPresentErrors": 0,
         "fieldInWrongPositionErrors": 0,
         "valueMismatchErrors": 0,
         "wrongFieldTypeErrors": 0}

error_found = False

for object in desired_output:
    for field in object.keys():
        
        path_to_field = [field]
        
        results, stats_dict, error_found = field_checker(results=results, 
                                                stats_dict=stats, 
                                                desired_field_key=field,
                                                path_to_field=path_to_field,
                                                desired_output=desired_output, 
                                                returned_output=returned_output, 
                                                error_found=error_found)
    
results_json = results.reset_index().to_json(orient='records')

print_to_screen = {
    "stats": stats_dict,
    "testResults": results_json
}

print(print_to_screen)
