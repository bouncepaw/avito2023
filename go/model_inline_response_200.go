/*
 * Customer segmentation
 *
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * API version: 1.0.0
 * Contact: bouncepaw2@ya.ru
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

type InlineResponse200 struct {
	// Status of the operation. If `ok`, then the operation went correctly, and you can ignore the `error` field. If `error`, an error occured which is specified in the `error` field.
	Status string `json:"status"`
	// Explanation of the error. Set only if `status` is `error`.  Values: * `name taken` means the provided name is taken already and cannot be used for new segments. * `name free` means that no segment with the given name exists. * Other values are internal errors.
	Error_ string `json:"error,omitempty"`
}
