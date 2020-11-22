// +build !ignore_autogenerated

/*
The License
*/
// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PassboltSecret) DeepCopyInto(out *PassboltSecret) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PassboltSecret.
func (in *PassboltSecret) DeepCopy() *PassboltSecret {
	if in == nil {
		return nil
	}
	out := new(PassboltSecret)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PassboltSecret) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PassboltSecretList) DeepCopyInto(out *PassboltSecretList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PassboltSecret, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PassboltSecretList.
func (in *PassboltSecretList) DeepCopy() *PassboltSecretList {
	if in == nil {
		return nil
	}
	out := new(PassboltSecretList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PassboltSecretList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PassboltSecretSource) DeepCopyInto(out *PassboltSecretSource) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PassboltSecretSource.
func (in *PassboltSecretSource) DeepCopy() *PassboltSecretSource {
	if in == nil {
		return nil
	}
	out := new(PassboltSecretSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PassboltSecretSpec) DeepCopyInto(out *PassboltSecretSpec) {
	*out = *in
	out.Source = in.Source
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PassboltSecretSpec.
func (in *PassboltSecretSpec) DeepCopy() *PassboltSecretSpec {
	if in == nil {
		return nil
	}
	out := new(PassboltSecretSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PassboltSecretStatus) DeepCopyInto(out *PassboltSecretStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PassboltSecretStatus.
func (in *PassboltSecretStatus) DeepCopy() *PassboltSecretStatus {
	if in == nil {
		return nil
	}
	out := new(PassboltSecretStatus)
	in.DeepCopyInto(out)
	return out
}
