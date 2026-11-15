export const load = ({ params }: { params: { name: string } }) => {
	return {
		packagePath: params.name
	};
};
